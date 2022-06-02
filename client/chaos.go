package client

import (
	"fmt"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/rs/zerolog/log"
)

// Chaos is controller that manages Chaosmesh CRD instances to run experiments
type Chaos struct {
	Client         *K8sClient
	ResourceByName map[string]string
	Namespace      string
}

// NewChaos creates controller to run and stop chaos experiments
func NewChaos(client *K8sClient, namespace string) *Chaos {
	return &Chaos{
		Client:         client,
		ResourceByName: make(map[string]string),
		Namespace:      namespace,
	}
}

// Run runs experiment and saves it's ID
func (c *Chaos) Run(app cdk8s.App, id string, resource string) (string, error) {
	log.Info().Msg("Applying chaos experiment")
	manifest := app.SynthYaml().(string)
	fmt.Println(manifest)
	c.ResourceByName[id] = resource
	if err := c.Client.Apply(manifest); err != nil {
		return id, err
	}
	return id, nil
}

// Stop removes a chaos experiment
func (c *Chaos) Stop(id string) error {
	defer delete(c.ResourceByName, id)
	return c.Client.DeleteResource(c.Namespace, c.ResourceByName[id], id)
}
