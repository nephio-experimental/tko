package server

import (
	contextpkg "context"
	"fmt"

	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
	backendpkg "github.com/nephio-experimental/tko/backend"
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

		Kind:        "Site",
		ListKind:    "SiteList",
		Singular:    "site",
		Plural:      "sites",
		ObjectTyper: Scheme,

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

		ListFunc: func(context contextpkg.Context, store *Store) (runtime.Object, error) {
			var krmSiteList krm.SiteList
			krmSiteList.APIVersion = APIVersion
			krmSiteList.Kind = "SiteList"

			if results, err := store.Backend.ListSites(context, backendpkg.ListSites{}); err == nil {
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

		TableFunc: func(context contextpkg.Context, store *Store, object runtime.Object, options *meta.TableOptions) (*meta.Table, error) {
			table := new(meta.Table)

			krmSites, err := ToSitesKRM(object)
			if err != nil {
				return nil, err
			}

			if (options == nil) || !options.NoHeaders {
				descriptions := krm.Site{}.TypeMeta.SwaggerDoc()
				nameDescription, _ := descriptions["name"]
				siteIdDescription, _ := descriptions["siteId"]
				templateIdDescription, _ := descriptions["templateId"]
				table.ColumnDefinitions = []meta.TableColumnDefinition{
					{Name: "Name", Type: "string", Format: "name", Description: nameDescription},
					{Name: "SiteID", Type: "string", Description: siteIdDescription},
					{Name: "TemplateID", Type: "string", Description: templateIdDescription},
					//{Name: "Metadata", Description: descriptions["metadata"]},
				}
			}

			table.Rows = make([]meta.TableRow, len(krmSites))
			for index, krmSite := range krmSites {
				row := meta.TableRow{
					Cells: []any{krmSite.Name, krmSite.Spec.SiteId, krmSite.Spec.TemplateId},
				}
				if (options == nil) || (options.IncludeObject != meta.IncludeNone) {
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
	name, err := IDToName(siteInfo.SiteID)
	if err != nil {
		return krm.Site{}, err
	}

	var krmSite krm.Site
	krmSite.APIVersion = APIVersion
	krmSite.Kind = "Site"
	krmSite.Name = name
	krmSite.UID = types.UID("tko|site|" + siteInfo.SiteID)

	if siteId := siteInfo.SiteID; siteId != "" {
		krmSite.Spec.SiteId = &siteId
	}
	if templateId := siteInfo.TemplateID; templateId != "" {
		krmSite.Spec.TemplateId = &templateId
	}
	krmSite.Spec.Metadata = siteInfo.Metadata
	krmSite.Spec.DeploymentIds = siteInfo.DeploymentIDs

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
	var id string
	if krmSite.Spec.SiteId != nil {
		id = *krmSite.Spec.SiteId
	}
	if id == "" {
		var err error
		if id, err = NameToID(krmSite.Name); err != nil {
			return nil, err
		}
	}

	site := backendpkg.Site{
		SiteInfo: backendpkg.SiteInfo{
			SiteID:   id,
			Metadata: krmSite.Spec.Metadata,
		},
	}

	if krmSite.Spec.TemplateId != nil {
		site.TemplateID = *krmSite.Spec.TemplateId
	}

	site.Resources = ResourcesFromKRM(krmSite.Spec.Package)

	return &site, nil
}
