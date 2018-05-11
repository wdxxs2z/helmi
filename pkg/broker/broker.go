package broker

import (
	"os"
	"fmt"
	"log"
	"strings"
	"context"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"

	"github.com/wdxxs2z/helmi/pkg/catalog"
	"github.com/wdxxs2z/helmi/pkg/release"
	"github.com/wdxxs2z/helmi/pkg/config"
	helmi "github.com/wdxxs2z/helmi/pkg/helm"
	helmicons "github.com/wdxxs2z/helmi/pkg/constants"
	"errors"
)

type UserCredentials map[string]interface{}

type ProvisionParameters map[string]string

type BindParameters map[string]interface{}

type UpdateParameters map[string]interface{}

type RequestContext map[string]string

type HelmBroker struct {
	allowUserProvisionParameters 	bool
	allowUserUpdateParameters    	bool
	allowUserBindParameters      	bool
	catalog 			catalog.Catalog
	logger                  	lager.Logger
	brokerRouter			*mux.Router
	helmClient                      *helmi.Client
}

type CatalogExternal struct {
	Services []brokerapi.Service `json:"services"`
}

func New(config config.Config, catalog catalog.Catalog, client *helmi.Client, logger lager.Logger) *HelmBroker {
	brokerRouter := mux.NewRouter()
	broker := &HelmBroker{
		allowUserProvisionParameters: 	config.AllowUserProvisionParameters,
		allowUserUpdateParameters:      config.AllowUserUpdateParameters,
		allowUserBindParameters:        config.AllowUserBindParameters,
		catalog:			catalog,
		logger:				logger.Session("service-broker"),
		brokerRouter:			brokerRouter,
		helmClient:                     client,
	}
	brokerapi.AttachRoutes(broker.brokerRouter, broker, logger)
	liveness := broker.brokerRouter.HandleFunc("/liveness", livenessHandler).Methods(http.MethodGet)

	broker.brokerRouter.Use(authHandler(config, map[*mux.Route]bool{liveness: true}))
	broker.brokerRouter.Use(handlers.ProxyHeaders)
	broker.brokerRouter.Use(handlers.CompressHandler)
	broker.brokerRouter.Use(handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{http.MethodHead, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}),
		handlers.AllowCredentials(),
	))

	return broker
}

func (b *HelmBroker) Run(address string) {
	if b.helmClient == nil {
		b.logger.Error("run-service-broker-error", errors.New("helm client must not null"), lager.Data{"advance": "check your kubeconfig, or catalog.yaml"})
		return
	}
	log.Println("Helm Service Broker started on port " + strings.TrimPrefix(address, ":"))
	log.Fatal(http.ListenAndServe(address, b.brokerRouter))
}

func (b *HelmBroker) Services(context context.Context) ([]brokerapi.Service, error) {
	b.logger.Debug(helmicons.FetchServiceCatalog, lager.Data{})

	services := make([]brokerapi.Service, 0, len(b.catalog.Services))
	for _, service := range b.catalog.Services {
		servicePlans := make([]brokerapi.ServicePlan, 0, len(service.Plans))
		for _, plan := range service.Plans {

			planCosts := make([]brokerapi.ServicePlanCost, 0 , len(plan.Metadata.Costs))
			for _, cost := range plan.Metadata.Costs {
				c := brokerapi.ServicePlanCost{
					Amount:    cost.Amount,
					Unit:      cost.Unit,
				}
				planCosts = append(planCosts, c)
			}
			p := brokerapi.ServicePlan{
				ID:		plan.Id,
				Name:           plan.Name,
				Description:    plan.Description,
				Free:           plan.Free,
				Bindable:       plan.Bindable,
				Metadata:       &brokerapi.ServicePlanMetadata{
					Bullets: plan.Metadata.Bullets,
					Costs:   planCosts,
				},
			}
			servicePlans = append(servicePlans, p)
		}
		s:= brokerapi.Service{
			ID:		 service.Id,
			Name:            service.Name,
			Description:     service.Description,
			Bindable:        service.Bindable,
			Tags:            service.Tags,
			PlanUpdatable:   service.PlanUpdateable,
			Metadata:        &brokerapi.ServiceMetadata{
				DisplayName:		service.Metadata.DisplayName,
				ImageUrl: 		service.Metadata.ImageUrl,
				SupportUrl:             service.Metadata.SupportUrl,
				ProviderDisplayName:    service.Metadata.ProviderDisplayName,
				DocumentationUrl:       service.Metadata.DocumentationUrl,
				LongDescription:        service.Metadata.LongDescription,
			},
			Plans:           servicePlans,
		}
		services = append(services, s)
	}
	return services, nil
}

func (b *HelmBroker) Provision(context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error){
	b.logger.Debug("provision", lager.Data{
		helmicons.InstanceIDLogKey:        	instanceID,
		helmicons.DetailsLogKey:           	details,
		helmicons.AcceptsIncompleteLogKey: 	asyncAllowed,
	})

	provisionParameters := ProvisionParameters{}
	if b.allowUserProvisionParameters && len(details.RawParameters) > 0 {
		if err:= json.Unmarshal(details.RawParameters, &provisionParameters); err!= nil {
			return brokerapi.ProvisionedServiceSpec{}, err
		}
	}

	servicePlan, err := b.catalog.GetServicePlan(details.ServiceID,details.PlanID)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("Service Plan '%s' not found", details.PlanID)
	}

	requestContext := RequestContext{}
	if len(details.RawContext) >0 {
		if err := json.Unmarshal(details.RawContext, & requestContext); err!= nil {
			return brokerapi.ProvisionedServiceSpec{}, err
		}
	}
	if err := release.Install(&b.catalog, details.ServiceID, servicePlan.Id, instanceID, asyncAllowed, provisionParameters, requestContext, b.helmClient, b.logger); err != nil {
		return brokerapi.ProvisionedServiceSpec{}, err
	}

	return brokerapi.ProvisionedServiceSpec{IsAsync: false}, nil
}

func (b *HelmBroker) Update(context context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	b.logger.Debug("update", lager.Data{
		helmicons.InstanceIDLogKey:        	instanceID,
		helmicons.DetailsLogKey:           	details,
		helmicons.AcceptsIncompleteLogKey: 	asyncAllowed,
	})

	// TODO

	return brokerapi.UpdateServiceSpec{IsAsync: false}, nil
}

func (b *HelmBroker) Deprovision(context context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	b.logger.Debug("deprovision", lager.Data{
		helmicons.InstanceIDLogKey:        	instanceID,
		helmicons.DetailsLogKey:           	details,
		helmicons.AcceptsIncompleteLogKey: 	asyncAllowed,
	})

	if err := release.Delete(instanceID, b.helmClient, b.logger); err != nil {
		return brokerapi.DeprovisionServiceSpec{}, err
	}

	return brokerapi.DeprovisionServiceSpec{IsAsync: false}, nil
}

func (b *HelmBroker) Bind(context context.Context, instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error){
	b.logger.Debug("bind", lager.Data{
		helmicons.InstanceIDLogKey: 		instanceID,
		helmicons.BindingIDLogKey:  	 	bindingID,
		helmicons.DetailsLogKey:       	details,
	})

	servicePlan, err := b.catalog.GetServicePlan(details.ServiceID,details.PlanID)
	if err != nil {
		return brokerapi.Binding{}, err
	}

	bindParameters := BindParameters{}
	if len(details.RawParameters) > 0 && b.allowUserBindParameters {
		if err := json.Unmarshal(details.RawParameters, &bindParameters); err != nil {
			return brokerapi.Binding{}, err
		}
	}

	credentials, err := release.GetCredentials(&b.catalog, details.ServiceID, servicePlan.Id, instanceID, b.helmClient, b.logger)
	if err != nil {
		return brokerapi.Binding{}, err
	}

	binding := brokerapi.Binding{Credentials: credentials}
	return binding, nil
}

func (b *HelmBroker) Unbind(context context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	b.logger.Debug("unbind", lager.Data{
		helmicons.InstanceIDLogKey: 		instanceID,
		helmicons.BindingIDLogKey:  	 	bindingID,
		helmicons.DetailsLogKey:       	details,
	})

	exists, err := release.Exists(instanceID, b.helmClient, b.logger)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return nil
}

func (b *HelmBroker) LastOperation(context context.Context, instanceID, operationData string) (brokerapi.LastOperation, error) {
	b.logger.Debug("last-operation", lager.Data{
		helmicons.InstanceIDLogKey: instanceID,
	})

	status, err := release.GetStatus(instanceID, b.helmClient, b.logger)
	if err != nil {
		return brokerapi.LastOperation{
			State: "failed",
		}, err
	}

	if status.IsFailed {
		return brokerapi.LastOperation{
			State: "failed",
		}, nil
	}

	if status.IsAvailable {
		return brokerapi.LastOperation{
			State: "success",
		},nil
	}

	if status.IsDeployed && !status.IsAvailable {
		return brokerapi.LastOperation{
			State: "in progress",
		}, nil
	}

	return brokerapi.LastOperation{},nil
}

func livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func authHandler(config config.Config, noAuthRequired map[*mux.Route]bool) mux.MiddlewareFunc{
	validCredentials := func(r *http.Request) bool {
		if noAuthRequired[mux.CurrentRoute(r)] {
			return true
		}
		user := os.Getenv("USERNAME")
		pass := os.Getenv("PASSWORD")
		username, password, ok := r.BasicAuth()
		if ok && username == user && password == pass {
			return true
		}
		return false
	}

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !validCredentials(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}
}