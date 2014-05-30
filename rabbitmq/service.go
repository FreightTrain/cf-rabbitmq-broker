// Copyright 2014, The cf-service-broker Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that
// can be found in the LICENSE file.

package rabbitmq

import (
	"fmt"
	"github.com/FreightTrain/cf-rabbitmq-broker/broker"
	"log"
)

// BrokerService implementation for RabbitMQ Server
type RabbitService struct {
	opts  ZoneOptions
	admin *rabbitAdmin
}

func New(opts ZoneOptions) (*RabbitService, error) {
	adm, err := getAdminClient(opts, opts.MgmtUser, opts.MgmtPass)
	if err != nil {
		return nil, err
	}
	return &RabbitService{opts, adm}, nil
}

func getAdminClient(opts ZoneOptions, username string, password string) (*rabbitAdmin, error) {
	url := fmt.Sprintf("http://%v:%v", opts.MgmtHost, opts.MgmtPort)
	adm, err := newRabbitAdmin(url, username, password)
	if err != nil {
		return nil, err
	}
	return adm, nil
}

func (b *RabbitService) Catalog() (broker.Catalog, error) {
	// TODO: Maybe read catalog from a file
	return broker.Catalog{
		Services: []broker.Service{
			broker.Service{
				Id:          "rabbitmq",
				Name:        "rabbitmq",
				Description: "RabbitMQ Message Broker",
				Bindable:    true,
				Tags:        []string{"rabbitmq", "messaging"},
				Plans: []broker.Plan{
					broker.Plan{
						Id:          "default",
						Name:        "default",
						Description: "Default RabbitMQ plan represented as a unique broker's vhost.",
					},
				},
			},
		},
	}, nil
}

func (b *RabbitService) Provision(pr broker.ProvisioningRequest) (string, error) {
	vhost := pr.InstanceId
	if err := b.admin.createVhost(vhost, false); err != nil {
		return "", err
	}
	log.Printf("Service: Virtual host created on %v: [%v]", b.admin.client.Endpoint, vhost)

	username := fmt.Sprintf("m-%v", vhost)
	password, _ := broker.RandomPasswordGenerator.GeneratePassword()
	if err := b.admin.createUser(username, password); err != nil {
		b.admin.deleteVhost(vhost)
		return "", err
	}
	log.Printf("Service: Management user created on %v: [%v:%v]", b.admin.client.Endpoint, username, password)

	if err := b.admin.grantAllPermissionsIn(username, vhost); err != nil {
		b.admin.deleteUser(username)
		b.admin.deleteVhost(vhost)
		return "", err
	}
	log.Printf("Service: All permissions granted to management user on %v: [%v]", b.admin.client.Endpoint, username)

	// Instantiate a new admin client against the other zone, using the mgmt user/pass created earlier
	mgmtClient, err := getAdminClient(b.opts, username, password)
	if err != nil {
		return "", err
	}

	for _, zoneOpts := range Opts.Zones {

		if zoneOpts.Name == b.opts.Name {
			continue
		}

		upstreamName := fmt.Sprintf("f-%v", vhost)
		upstreamOpts := map[string]interface{}{
			"uri":            fmt.Sprintf("amqp://%v:%v@%v:%v/%v", zoneOpts.MgmtUser, zoneOpts.MgmtPass, zoneOpts.Host, zoneOpts.Port, vhost),
			"maxHops":        1,
			"expires":        36000000,
			"reconnectDelay": 5,
			"ackMode":        "on-confirm",
			"prefetchCount":  1,
		}
		if err := mgmtClient.setFederationUpstream(vhost, upstreamName, upstreamOpts); err != nil {
		}
		log.Printf("Service: Federation Upstream set for %v@%v", vhost, upstreamName)

		policyName := fmt.Sprintf("p-%v", vhost)
		policyOpts := map[string]interface{}{
			"vhost":    vhost,
			"pattern":  "^ps\\.",
			"applyTo":  "all",
			"name":     policyName,
			"priority": 0,
			"definition": map[string]interface{}{
				"federation-upstream-set": "all",
			},
		}
		if err := mgmtClient.setFederationPolicy(vhost, policyName, policyOpts); err != nil {
		}
		log.Printf("Service: Federation Policy set for %v@%v", vhost, upstreamName)

	}

	dashboardUrl := fmt.Sprintf("http://%v:%v/#/login/%v/%v", b.opts.Host, b.opts.MgmtPort, username, password)
	log.Printf("Service: Dashboard URL generated: [%v]", dashboardUrl)

	return dashboardUrl, nil
}

func (b *RabbitService) Deprovision(pr broker.ProvisioningRequest) error {
	vhost := pr.InstanceId
	username := fmt.Sprintf("m-%v", vhost)
	if err := b.admin.deleteUser(username); err != nil {
		return err
	}
	log.Printf("Service: Management user deleted: [%v]", username)

	//TODO:Should close existing connections from user 'username'???

	if err := b.admin.deleteVhost(vhost); err != nil {
		return err
	}
	log.Printf("Service: Virtual host deleted: [%v]", vhost)

	return nil
}

func (b *RabbitService) Bind(br broker.BindingRequest) (string, broker.Credentials, string, error) {
	vhost := br.InstanceId

	username := fmt.Sprintf("u-%v", vhost)
	password, _ := broker.RandomPasswordGenerator.GeneratePassword()
	if err := b.admin.createUser(username, password); err != nil {
		return "", nil, "", err
	}
	log.Printf("Service: User created: [%v]", username)

	if err := b.admin.grantAllPermissionsIn(username, vhost); err != nil {
		b.admin.deleteUser(username)
		return "", nil, "", err
	}
	log.Printf("Service: All permissions granted for vhost: [%v] to user: [%v]", vhost, username)

	amqpUrl := fmt.Sprintf("amqp://%v:%v@%v:%v/%v", username, password, b.opts.Host, b.opts.Port, vhost)
	log.Printf("Service: AMQP URL generated: [%v]", amqpUrl)

	return b.opts.Name, broker.Credentials{
		"uri":  amqpUrl,
		"host": b.opts.Host,
	}, "", nil
}

func (b *RabbitService) Unbind(br broker.BindingRequest) error {
	vhost := br.InstanceId
	username := fmt.Sprintf("u-%v", vhost)

	log.Printf("Service: Deleting user: [%v]", username)

	err := b.admin.deleteUser(username)
	if err != nil {
		return err
	}
	log.Printf("Service: User deleted: [%v]", username)

	//TODO:Should close existing connections from user 'username'???

	return nil
}
