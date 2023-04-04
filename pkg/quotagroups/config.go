package quotagroups

import (
	"context"
	"encoding/json"
	pApi "github.com/openshift/api/project/v1"
	qApi "github.com/openshift/api/quota/v1"
	projectV1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	quota "github.com/openshift/client-go/quota/clientset/versioned/typed/quota/v1"
	quotaV1 "github.com/openshift/client-go/quota/clientset/versioned/typed/quota/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultTimeOut int64 = 10
)

type Config struct {
	Labels             []string `yaml:"labels"`
	labels             map[string]bool
	CatchAllQuotaGroup string `yaml:"catch_all_quota_group"`
	PatchPath          string `yaml:"patch_path"`
	blockedQuotaGroups groups
	TimeOut            int64 `yaml:"time_out"`
}

func (c *Config) Initialize() {
	if c.TimeOut <= 1 {
		timeOut := defaultTimeOut
		c.TimeOut = timeOut
	}
	if len(c.Labels) == 0 {
		c.Labels = []string{"quotagroup"}
	}
	labels := make(map[string]bool)
	for _, label := range c.Labels {
		labels[label] = true
	}
	c.labels = labels
	if c.CatchAllQuotaGroup == "" {
		c.CatchAllQuotaGroup = "blockedresources"
	}
	if c.PatchPath == "" {
		c.PatchPath = "/spec/selector/labels/matchExpressions/0/values/-"
	}
}

// patchStringValue specifies a patch operation for a string.
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (c *Config) setBlockedQuotaGroups(quota qApi.ClusterResourceQuota) {
	if quota.Name != c.CatchAllQuotaGroup {
		return
	}
	log.Debugf("initializing groups listed in %s", c.CatchAllQuotaGroup)
	blockedQuotaGroups := make(groups)
	///spec/selector/labels/matchExpressions/0/values/-
	log.Debugf("looping though matchexpressions for %s", c.CatchAllQuotaGroup)
	for _, matchExpression := range quota.Spec.Selector.LabelSelector.MatchExpressions {
		if _, exists := c.labels[matchExpression.Key]; exists {
			log.Debugf("looping though matchexpression values for %s", c.CatchAllQuotaGroup)
			for _, blockedQuotaGroup := range matchExpression.Values {
				log.Debugf("adding %s to the internal list of blockedQuotaGroups", c.CatchAllQuotaGroup)
				blockedQuotaGroups[blockedQuotaGroup] = true
			}
		}
	}
	c.blockedQuotaGroups = blockedQuotaGroups
}

func (c *Config) getQuotaGroups() (groups, error) {
	//if config, err := versioned.NewForConfig(nil); err != nil {
	//	panic(err)
	//} else {
	quotaGroups := make(groups)
	if quotaClient, configErr := quotaV1.NewForConfig(getKubeConfig()); configErr != nil {
		return nil, configErr
	} else if quotas, listErr := quotaClient.ClusterResourceQuotas().List(ctx, v1.ListOptions{}); listErr != nil {
		log.Errorf("Error while retrieving all projects: %e", listErr)
		return nil, listErr
	} else {
		var q qApi.ClusterResourceQuota
		for _, q = range quotas.Items {
			log.Debug("adding to internal list of quota groups")
			quotaGroups[q.Name] = true
			c.setBlockedQuotaGroups(q)
		}
	}
	log.Debugf("list of groups: %v", quotaGroups)
	//informer := informers.NewClusterResourceQuotaInformer(config, time.Minute, cache.Indexers{})
	//informer.GetController().Run()
	//}
	return quotaGroups, nil
}

func (c Config) GetProjectLabel(labels map[string]string) string {
	for _, label := range c.Labels {
		if group, exists := labels[label]; exists {
			return group
		}
	}
	return ""
}

func (c Config) getProjectGroups() (groups, error) {
	log.Debug("getting a list of projects")
	projectGroups := make(groups)
	if projectClient, err := projectV1.NewForConfig(getKubeConfig()); err != nil {
		return nil, err
	} else if projects, listErr := projectClient.Projects().List(ctx, v1.ListOptions{
		TimeoutSeconds: &(c.TimeOut),
		//ResourceVersionMatch: v1.ResourceVersionMatchNotOlderThan,
		TypeMeta: v1.TypeMeta{
			Kind:       "Project",
			APIVersion: "project.openshift.io/v1",
		},
	}); listErr != nil {
		log.Errorf("Error while retrieving all projects: %e", listErr)
		return nil, err
	} else {
		log.Debug("getting a list of QuotaGroups")

		var p pApi.Project
		log.Debugf("looping through %d projects", projects.Size())
		for _, p = range projects.Items {
			projectGroups[c.GetProjectLabel(p.GetLabels())] = true
		}
	}
	return projectGroups, nil
}

func (c Config) PatchProjectQuotaGroups() error {
	if patchClient, clientErr := quota.NewForConfig(getKubeConfig()); clientErr != nil {
		log.Errorf("Error while connecting to OpenShift: %e", clientErr)
		return clientErr
	} else if quotaGroups, quotaGroupsErr := c.getQuotaGroups(); quotaGroupsErr != nil {
		return quotaGroupsErr
	} else if projectGroups, projectGroupsErr := c.getProjectGroups(); projectGroupsErr != nil {
		return projectGroupsErr
	} else {
		for invalidGroup := range projectGroups.Difference(quotaGroups).Difference(c.blockedQuotaGroups) {
			log.Infof("ClusterResourceQuota %s neither exists nor is it blocked. Patching %s to have it blocked",
				invalidGroup, c.CatchAllQuotaGroup)
			payload := []patchStringValue{{
				Op:    "add",
				Path:  c.PatchPath,
				Value: invalidGroup,
			}}
			log.Debug("marshalling patch payload to json")
			if payloadBytes, marshalErr := json.Marshal(payload); marshalErr != nil {
				log.Errorf("Error while marshalling patch payload to json: %e", marshalErr)
				return marshalErr
			} else if result, patchErr := patchClient.ClusterResourceQuotas().Patch(context.Background(), c.CatchAllQuotaGroup,
				types.JSONPatchType, payloadBytes, v1.PatchOptions{}); patchErr != nil {
				return patchErr
			} else {
				log.Debugf("%s succesfully patched, payload: %s", c.CatchAllQuotaGroup, payloadBytes)
				log.Debugf("%v", result)
			}
		}
	}
	log.Debug("Done")

	return nil
}
