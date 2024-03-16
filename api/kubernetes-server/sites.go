package server

import (
	contextpkg "context"
	"fmt"
	"time"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
	backendpkg "github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewSiteStore(backend backend.Backend, log commonlog.Logger) *Store {
	store := Store{
		Backend: backend,
		Log:     log,

		TypeKind:          "Site",
		TypeListKind:      "SiteList",
		TypeSingular:      "site",
		TypePlural:        "sites",
		CanCreateOnUpdate: true,
		ObjectTyper:       Scheme,

		NewObjectFunc: func() runtime.Object {
			return new(krm.Site)
		},

		NewListObjectFunc: func() runtime.Object {
			return new(krm.SiteList)
		},

		GetFieldsFunc: func(object runtime.Object) (fields.Set, error) {
			if krmSite, ok := object.(*krm.Site); ok {
				fields := fields.Set{
					"metadata.name": krmSite.Name,
				}
				if krmSite.Spec.SiteId != nil {
					fields["spec.siteId"] = *krmSite.Spec.SiteId
				}
				if krmSite.Spec.TemplateId != nil {
					fields["spec.templateId"] = *krmSite.Spec.TemplateId
				}
				return fields, nil
			} else {
				return nil, fmt.Errorf("not a site: %T", object)
			}
		},

		CreateFunc: func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error) {
			if site, err := SiteFromKRM(object); err == nil {
				if err := store.Backend.SetSite(context, site); err == nil {
					return object, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		DeleteFunc: func(context contextpkg.Context, store *Store, id string) error {
			return store.Backend.DeleteSite(context, id)
		},

		PurgeFunc: func(context contextpkg.Context, store *Store) error {
			return store.Backend.PurgePlugins(context, backendpkg.SelectPlugins{})
		},

		GetFunc: func(context contextpkg.Context, store *Store, id string) (runtime.Object, error) {
			if site, err := store.Backend.GetSite(context, id); err == nil {
				if krmSite, err := SiteToKRM(site); err == nil {
					return krmSite, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		ListFunc: func(context contextpkg.Context, store *Store, options *metainternalversion.ListOptions, offset uint, maxCount uint) (runtime.Object, error) {
			var krmSiteList krm.SiteList

			idPatterns, err := IDPatternsFromListOptions(options)
			if err != nil {
				return nil, err
			}
			metadataPatterns, err := MetadataPatternsFromListOptions(options)
			if err != nil {
				return nil, err
			}

			if results, err := store.Backend.ListSites(context, backendpkg.SelectSites{SiteIDPatterns: idPatterns, MetadataPatterns: metadataPatterns}, backendpkg.Window{Offset: offset, MaxCount: int(maxCount)}); err == nil {
				if err := util.IterateResults(results, func(siteInfo backendpkg.SiteInfo) error {
					if krmSite, err := SiteInfoToKRM(&siteInfo); err == nil {
						krmSiteList.Items = append(krmSiteList.Items, *krmSite)
						return nil
					} else {
						return err
					}
				}); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}

			krmSiteList.APIVersion = APIVersion
			krmSiteList.Kind = "SiteList"
			return &krmSiteList, nil
		},

		TableFunc: func(context contextpkg.Context, store *Store, object runtime.Object, withHeaders bool, withObject bool) (*meta.Table, error) {
			table := new(meta.Table)

			krmSites, err := ToSitesKRM(object)
			if err != nil {
				return nil, err
			}

			if withHeaders {
				table.ColumnDefinitions = []meta.TableColumnDefinition{
					{Name: "Name", Type: "string", Format: "name"},
					{Name: "SiteID", Type: "string"},
					{Name: "TemplateID", Type: "string"},
					{Name: "Updated", Type: "string", Format: "date-time"},
				}
			}

			table.Rows = make([]meta.TableRow, len(krmSites))
			for index, krmSite := range krmSites {
				var updated time.Time
				var err error
				if updated, err = FromResourceVersion(krmSite.ResourceVersion); err != nil {
					return nil, err
				}

				row := meta.TableRow{
					Cells: []any{
						krmSite.Name,
						krmSite.Spec.SiteId,
						krmSite.Spec.TemplateId,
						updated,
					},
				}
				if withObject {
					row.Object = runtime.RawExtension{Object: &krmSite}
				}
				table.Rows[index] = row
			}

			return table, nil
		},
	}

	store.Init()
	return &store
}

func ToSitesKRM(object runtime.Object) ([]krm.Site, error) {
	switch object_ := object.(type) {
	case *krm.SiteList:
		return object_.Items, nil
	case *krm.Site:
		return []krm.Site{*object_}, nil
	default:
		return nil, backendpkg.NewBadArgumentErrorf("unsupported type: %T", object)
	}
}

func SiteInfoToKRM(siteInfo *backendpkg.SiteInfo) (*krm.Site, error) {
	name, err := tkoutil.ToKubernetesName(siteInfo.SiteID)
	if err != nil {
		return nil, backendpkg.NewBadArgumentError(err.Error())
	}

	var krmSite krm.Site
	krmSite.APIVersion = APIVersion
	krmSite.Kind = "Site"
	krmSite.Name = name
	krmSite.UID = types.UID("tko|site|" + siteInfo.SiteID)
	krmSite.ResourceVersion = ToResourceVersion(siteInfo.Updated)
	krmSite.Labels, _ = tkoutil.ToKubernetesNames(siteInfo.Metadata)

	siteId := siteInfo.SiteID
	krmSite.Spec.SiteId = &siteId
	if templateId := siteInfo.TemplateID; templateId != "" {
		krmSite.Spec.TemplateId = &templateId
	}
	krmSite.Status.DeploymentIds = siteInfo.DeploymentIDs

	return &krmSite, nil
}

func SiteToKRM(site *backendpkg.Site) (*krm.Site, error) {
	if krmSite, err := SiteInfoToKRM(&site.SiteInfo); err == nil {
		krmSite.Spec.Package = PackageToKRM(site.Package)
		return krmSite, nil
	} else {
		return nil, err
	}
}

func SiteFromKRM(object runtime.Object) (*backendpkg.Site, error) {
	var krmSite *krm.Site
	var ok bool
	if krmSite, ok = object.(*krm.Site); !ok {
		return nil, backendpkg.NewBadArgumentErrorf("not a Site: %T", object)
	}

	var siteId string
	var err error
	if siteId, err = tkoutil.FromKubernetesName(krmSite.Name); err != nil {
		return nil, backendpkg.NewBadArgumentError(err.Error())
	}

	var updated time.Time
	if updated, err = FromResourceVersion(krmSite.ResourceVersion); err != nil {
		return nil, err
	}

	site := backendpkg.Site{
		SiteInfo: backendpkg.SiteInfo{
			SiteID:  siteId,
			Updated: updated,
		},
	}

	site.Metadata, _ = tkoutil.FromKubernetesNames(krmSite.Labels)
	site.Package = PackageFromKRM(krmSite.Spec.Package)

	if krmSite.Spec.TemplateId != nil {
		site.TemplateID = *krmSite.Spec.TemplateId
	}

	return &site, nil
}
