package templateinstance

import (
	"context"
	"fmt"
	"encoding/json"

	tmaxv1 "template-instance-operator/pkg/apis/tmax/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crdapi "github.com/kubernetes-client/go/kubernetes/client"   
	"github.com/kubernetes-client/go/kubernetes/config"

	"github.com/tidwall/gjson"
)

var log = logf.Log.WithName("controller_templateinstance")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TemplateInstance Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTemplateInstance{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("templateinstance-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TemplateInstance
	err = c.Watch(&source.Kind{Type: &tmaxv1.TemplateInstance{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TemplateInstance
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TemplateInstance{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileTemplateInstance implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTemplateInstance{}

// ReconcileTemplateInstance reconciles a TemplateInstance object
type ReconcileTemplateInstance struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Template CR Type
type TemplateCR struct {
	spec TemplateCRSpec
}

// Detail Template CR Spec
type TemplateCRSpec struct {
	kind string
	labels []byte
	operatorStartTime string
	shortDescription string
	longDescription string
	provider string
	imageUrl string
	recommend bool
	tags []string
	objectKinds []string
	metadata []byte
	plans []byte
	objects []ObjectSpec
	parameters []byte
}

type ObjectSpec struct {
	kind string
	fields []byte
}

// Reconcile reads that state of the cluster for a TemplateInstance object and makes changes based on the state read
// and what is in the TemplateInstance.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTemplateInstance) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TemplateInstance")

	// Fetch the TemplateInstance instance
	instance := &tmaxv1.TemplateInstance{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Template instance

	templateNameSpace := request.Namespace
	templateName := instance.Spec.Template.Metadata.Name

	c, err := config.LoadKubeConfig()
	if err != nil {
		return reconcile.Result{},err
	}
	clientset := crdapi.NewAPIClient(c)
	
	// get template cr's info
	templateCR,_,err := clientset.CustomObjectsApi.GetNamespacedCustomObject(context.Background(),"tmax.io","v1",templateNameSpace,"templates",templateName);
	if err != nil {
		panic("===[ Template Error ] : " + err.Error())
	}

	// map[string]interface{} to []byte
	convert, err := json.Marshal(templateCR)
	if err != nil {
		panic("===[ Marshal Error ] : " + err.Error())
	}

	// test gjson
	value := gjson.Get(string(convert), "spec.objects")
	reqLogger.Info("************ spec.objects : " + value.String())

	// []byte to interface{}
	var tcr TemplateCR
	err = json.Unmarshal(convert, &tcr)
	if err != nil {
		panic("===[ Unmarshal Error ] : " + err.Error())
	}

	// push instance parameters into template cr's object
	// for _, parameter := range instance.Spec.Template.Parameters {
	// 	//
	// }

	// deploy template cr's object 
	for _, object := range tcr.spec.objects {
		err = deploy(object)
		if err != nil {
			panic("===[ Deploy Error ] : " + err.Error())
		}
	}

	return reconcile.Result{}, nil
}

// Error Wrapping
type DeployError struct {
	kind string
}

func (de *DeployError) Error() string {
	return fmt.Sprint("It is not a type of kind %v", de.kind)
}

// classify object
func deploy(object interface{}) error {
	// TODO : 모든 kind를 일반화 해서 apply 하는 방법...?

	objectJson, err := json.Marshal(object)
	if err != nil {
		panic("===[ Marshal Error ] : " + err.Error())
	}

	kind := gjson.Get(string(objectJson), "kind")

	switch kind.String() {
	case "Service":
		deployInstance(object, "services")
		return nil
	case "Deployment":
		deployInstance(object, "deployments")
		return nil
	case "Pod":
		deployInstance(object, "pods")
		return nil
	default:
		return &DeployError{kind.String()}
	}
	return nil
}

// apply instance
func deployInstance(object interface{}, plural string) error {
	objectJson, err := json.Marshal(object)
	if err != nil {
		panic("===[ Marshal Error ] : " + err.Error())
	}

	group := gjson.Get(string(objectJson), "fields.apiVersion") // ?
	version := gjson.Get(string(objectJson), "fields.apiVersion")
	namespaceValue := gjson.Get(string(objectJson), "fields.metadata.namespace")

	var namespace string
	if len(namespaceValue.String()) != 0 {
		namespace = namespaceValue.String()
	} else {
		namespace = "default"
	}

	conf, err := config.LoadKubeConfig()
	if err != nil {
		return err
	}

	clientSet := crdapi.NewAPIClient(conf)
	response, _, err := clientSet.CustomObjectsApi.CreateNamespacedCustomObject(context.Background(), group.String(), version.String(), namespace, plural, object, nil)

	if err != nil && response == nil {
		if errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}
