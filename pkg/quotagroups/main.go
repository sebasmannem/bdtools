package quotagroups

import (
	"context"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	log        *zap.SugaredLogger
	ctx        context.Context
	ValidRoles = map[string]bool{
		"primary": true,
		"standby": true,
	}
)

func InitLogger(logger *zap.SugaredLogger) {
	log = logger
}

func InitContext(c context.Context) {
	ctx = c
}

func getKubeConfig() *rest.Config {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here

	configOverrides := &clientcmd.ConfigOverrides{}
	// if you want to change override values or bind them to flags, there are methods to help you

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	if ns, overridden, err := kubeConfig.Namespace(); err != nil {
		log.Panicf("error occurred while retrievinf namepsace config")
	} else {
		var sOverride string
		if !overridden {
			sOverride = "not "
		}
		log.Debugf("namespace is %s and it is %soverridden", ns, sOverride)
	}

	if config, err := kubeConfig.ClientConfig(); err != nil {
		log.Panicf("I could not retrieve openshift connection config: %e", err)
	} else {
		log.Debugf("config for connecting as %s @ %s", config.Username, config.Host)
		return config
	}
	return nil
}
