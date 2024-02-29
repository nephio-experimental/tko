package server

import (
	contextpkg "context"
	"fmt"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
	backendpkg "github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewSiteStore(backend backend.Backend, log commonlog.Logger) *Store {
	store := Store{
		Backend: backend,
		Log:     log,

		TypeKind:     "Site",
		TypeListKind: "SiteList",
		TypeSingular: "site",
		TypePlural:   "sites",
		ObjectTyper:  Scheme,

		NewResourceFunc: func() runtime.Object {
			return new(krm.Site)
		},

		NewResourceListFunc: func() runtime.Object {
			return new(krm.SiteList)
		},

		CreateFunc: func(context contextpkg.Context, store *Store, object runtime.Object) (runtime.Object, error) {
			if krmSite, ok := object.(*krm.Site); ok {
				if site, err := KRMToSite(krmSite); err == nil {
					if err := store.Backend.SetSite(context, site); err == nil {
						return krmSite, nil
					} else {
						return nil, err
					}
				} else {
					return nil, backendpkg.NewBadArgumentError(err.Error())
				}
			} else {
				return nil, backendpkg.NewBadArgumentErrorf("not a Site: %T", object)
			}
		},

		DeleteFunc: func(context contextpkg.Context, store *Store, id string) error {
			return store.Backend.DeleteSite(context, id)
		},

		GetFunc: func(context contextpkg.Context, store *Store, id string) (runtime.Object, error) {
			if site, err := store.Backend.GetSite(context, id); err == nil {
				if krmSite, err := SiteToKRM(site); err == nil {
					return &krmSite, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		},

		ListFunc: func(context contextpkg.Context, store *Store, offset uint, maxCount uint) (runtime.Object, error) {
			var krmSiteList krm.SiteList
			krmSiteList.APIVersion = APIVersion
			krmSiteList.Kind = "SiteList"

			if results, err := store.Backend.ListSites(context, backendpkg.ListSites{Offset: offset, MaxCount: maxCount}); err == nil {
				if err := util.IterateResults(results, func(siteInfo backendpkg.SiteInfo) error {
					if krmSite, err := SiteInfoToKRM(&siteInfo); err == nil {
						krmSiteList.Items = append(krmSiteList.Items, krmSite)
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
				}
			}

			table.Rows = make([]meta.TableRow, len(krmSites))
			for index, krmSite := range krmSites {
				row := meta.TableRow{
					Cells: []any{
						krmSite.Name,
						krmSite.Spec.SiteId,
						krmSite.Spec.TemplateId,
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
		return nil, fmt.Errorf("unsupported type: %T", object)
	}
}

func SiteInfoToKRM(siteInfo *backendpkg.SiteInfo) (krm.Site, error) {
	name, err := tkoutil.ToKubernetesName(siteInfo.SiteID)
	if err != nil {
		return krm.Site{}, err
	}

	var krmSite krm.Site
	krmSite.APIVersion = APIVersion
	krmSite.Kind = "Site"
	krmSite.Name = name
	krmSite.UID = types.UID("tko|site|" + siteInfo.SiteID)

	siteId := siteInfo.SiteID
	krmSite.Spec.SiteId = &siteId
	if templateId := siteInfo.TemplateID; templateId != "" {
		krmSite.Spec.TemplateId = &templateId
	}
	krmSite.Spec.Metadata = siteInfo.Metadata
	krmSite.Status.DeploymentIds = siteInfo.DeploymentIDs

	return krmSite, nil
}

func SiteToKRM(site *backendpkg.Site) (krm.Site, error) {
	if krmSite, err := SiteInfoToKRM(&site.SiteInfo); err == nil {
		krmSite.Spec.Package = ResourcesToKRM(site.Resources)
		return krmSite, nil
	} else {
		return krm.Site{}, err
	}
}

func KRMToSite(krmSite *krm.Site) (*backendpkg.Site, error) {
	var siteId string
	var err error
	if siteId, err = tkoutil.FromKubernetesName(krmSite.Name); err != nil {
		return nil, err
	}

	site := backendpkg.Site{
		SiteInfo: backendpkg.SiteInfo{
			SiteID:   siteId,
			Metadata: krmSite.Spec.Metadata,
		},
	}

	if krmSite.Spec.TemplateId != nil {
		site.TemplateID = *krmSite.Spec.TemplateId
	}

	site.Resources = ResourcesFromKRM(krmSite.Spec.Package)

	return &site, nil
}
