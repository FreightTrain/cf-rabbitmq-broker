// Copyright 2014, The cf-service-broker Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that
// can be found in the LICENSE file.

package rabbitmq

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/nimbus-cloud/cf-rabbitmq-broker/broker"
	"github.com/nimbus-cloud/rabbit-hole"
	"net/http"
)

type rabbitAdminError struct {
	code int
	err  error
}

func (e *rabbitAdminError) Code() int {
	return e.code
}
func (e *rabbitAdminError) Error() string {
	return fmt.Sprintf("%v: %v", e.code, e.err.Error())
}

type rabbitAdmin struct {
	client *rabbithole.Client
}

func newRabbitAdmin(brokerUrl, username, password string) (*rabbitAdmin, error) {
	client, err := rabbithole.NewClient(brokerUrl, username, password)
	if err != nil {
		return nil, err
	}
	return &rabbitAdmin{client}, nil
}

func (a *rabbitAdmin) isVhost(username string) (bool, error) {
	info, err := a.client.GetVhost(username)
	if info != nil {
		return true, nil
	} else if err.Error() == "not found" { // TODO: Create PR to expose the 404 in more user-friendly way
		return false, nil
	}
	return false, &rabbitAdminError{broker.ErrCodeOther, err}
}

func (a *rabbitAdmin) createVhost(vhostname string, tracing bool) error {
	if found, err := a.isVhost(vhostname); err != nil {
		return err
	} else if found {
		msg := fmt.Sprintf("Virtual host already exists: [%v]", vhostname)
		return &rabbitAdminError{broker.ErrCodeConflict, errors.New(msg)}
	}

	settings := rabbithole.VhostSettings{tracing}
	resp, err := a.client.PutVhost(vhostname, settings)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) deleteVhost(vhostname string) error {
	resp, err := a.client.DeleteVhost(vhostname)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) isUser(username string) (bool, error) {
	info, err := a.client.GetUser(username)
	if info != nil {
		return true, nil
	} else if err.Error() == "not found" {
		return false, nil
	}
	return false, &rabbitAdminError{broker.ErrCodeOther, err}
}

func (a *rabbitAdmin) createUser(username, password string) error {
	if found, err := a.isUser(username); err != nil {
		return err
	} else if found {
		msg := fmt.Sprintf("User already exists: %v", username)
		return &rabbitAdminError{broker.ErrCodeConflict, errors.New(msg)}
	}

	settings := rabbithole.UserSettings{
		Name:     username,
		Password: password,
		Tags:     "management, policymaker, monitoring",
	}
	resp, err := a.client.PutUser(username, settings)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) deleteUser(username string) error {
	resp, err := a.client.DeleteUser(username)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) grantAllPermissionsIn(username, vhostname string) error {
	unlimited := rabbithole.Permissions{".*", ".*", ".*"}
	resp, err := a.client.UpdatePermissionsIn(vhostname, username, unlimited)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) setFederationUpstream(vhost string, upstreamName string, fOpts map[string]interface{}) error {
	var fDef rabbithole.FederationDefinition
	err := mapstructure.Decode(fOpts, &fDef)
	resp, err := a.client.PutFederationUpstream(vhost, upstreamName, fDef)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) setFederationPolicy(vhost string, policyName string, pOpts map[string]interface{}) error {
	var pDef rabbithole.Policy
	err := mapstructure.Decode(pOpts, &pDef)
	resp, err := a.client.PutPolicy(vhost, policyName, pDef)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func checkResponseAndClose(resp *http.Response) error {
	defer resp.Body.Close()

	switch code := resp.StatusCode; code {
	case http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		err := errors.New("Entity not found")
		return &rabbitAdminError{broker.ErrCodeGone, err}
	default:
		err := errors.New(fmt.Sprintf("Unexpected response received: [%v]", code))
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
}
