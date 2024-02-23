package server

import (
	contextpkg "context"

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
			var siteList krm.SiteList
			siteList.APIVersion = APIVersion
			siteList.Kind = "SiteList"

			if results, err := store.Backend.ListSites(context, backendpkg.ListSites{}); err == nil {
				if err := util.IterateResults(results, func(siteInfo backendpkg.SiteInfo) error {
					if krmSite, err := SiteInfoToKRM(&siteInfo); err == nil {
						siteList.Items = append(siteList.Items, krmSite)
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

			return &siteList, nil
		},
	}

	store.Init()
	return &store
}

func SiteInfoToKRM(siteInfo *backendpkg.SiteInfo) (krm.Site, error) {
	var site krm.Site

	name, err := IDToName(siteInfo.SiteID)
	if err != nil {
		return site, err
	}

	site.APIVersion = APIVersion
	site.Kind = "Site"
	site.Name = name
	//site.GenerateName = "tko-site-"
	site.UID = types.UID("tko|site|" + siteInfo.SiteID)
	//site.ResourceVersion = "123"
	site.CreationTimestamp = meta.Now()

	siteId := siteInfo.SiteID
	site.Spec.SiteId = &siteId
	site.Spec.Metadata = siteInfo.Metadata

	return site, nil
}

func SiteToKRM(site *backendpkg.Site) (krm.Site, error) {
	return SiteInfoToKRM(&site.SiteInfo)
}

func KRMToSite(site *krm.Site) (*backendpkg.Site, error) {
	metadata := site.Spec.Metadata

	var id string
	if site.Spec.SiteId != nil {
		id = *site.Spec.SiteId
	}
	if id == "" {
		var err error
		if id, err = NameToID(site.Name); err != nil {
			return nil, err
		}
	}

	return &backendpkg.Site{
		SiteInfo: backendpkg.SiteInfo{
			SiteID:   id,
			Metadata: metadata,
		},
	}, nil
}
