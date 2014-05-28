// Copyright 2014, The cf-service-broker Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that
// can be found in the LICENSE file.

package broker

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var empty struct{} = struct{}{}

type handler struct {
	brokerServices []BrokerService
}

func newHandler(bs []BrokerService) *handler {
	return &handler{bs}
}

func (h *handler) catalog(r *http.Request) responseEntity {

	log.Printf("Handler: Requesting catalog")

	if cat, err := h.brokerServices[0].Catalog(); err != nil {
		return handleServiceError(err)
	} else {
		log.Printf("Handler: Catalog retrieved")

		return responseEntity{http.StatusOK, cat}
	}
}

func (h *handler) provision(req *http.Request) responseEntity {
	vars := mux.Vars(req)
	preq := ProvisioningRequest{InstanceId: vars[instanceId]}

	log.Printf("Handler: Provisioning: %v", preq)

	if err := json.NewDecoder(req.Body).Decode(&preq); err != nil {
		handleDecodingError(err)
	}

	log.Printf("Handler: Provisioning request decoded: %v", preq)

	var url string
	var err error

	for _, brokerService := range h.brokerServices {
		url, err = brokerService.Provision(preq)
		if err != nil {
			return handleServiceError(err)
		}
	}

	log.Printf("Handler: Provisioned: %v", preq)

	return responseEntity{http.StatusCreated, struct {
		DashboardUrl string `json:"dashboard_url"`
	}{url}}
}

func (h *handler) deprovision(req *http.Request) responseEntity {
	vars := mux.Vars(req)
	preq := ProvisioningRequest{InstanceId: vars[instanceId]}

	log.Printf("Handler: Deprovisioning: %v", preq)

	for _, brokerService := range h.brokerServices {
		if err := brokerService.Deprovision(preq); err != nil {
			return handleServiceError(err)
		}
	}

	log.Printf("Handler: Deprovisioned: %v", preq)

	return responseEntity{http.StatusOK, empty}
}

func (h *handler) bind(req *http.Request) responseEntity {
	vars := mux.Vars(req)
	breq := BindingRequest{InstanceId: vars[instanceId], BindingId: vars[bindingId]}

	log.Printf("Handler: Binding: %v", breq)

	if err := json.NewDecoder(req.Body).Decode(&breq); err != nil {
		handleDecodingError(err)
	}

	log.Printf("Handler: Binding request decoded: %v", breq)

	zoneCreds := make(map[string]Credentials)
	var zone string
	var url string
	var cred Credentials
	var err error

	for _, brokerService := range h.brokerServices {
		zone, cred, url, err = brokerService.Bind(breq)
		if err != nil {
			return handleServiceError(err)
		}
		zoneCreds[zone] = cred
	}

	log.Printf("Handler: Bound: %v", breq)

	return responseEntity{http.StatusCreated, struct {
		Credentials    interface{} `json:"credentials"`
		SyslogDrainUrl string      `json:"syslog_drain_url "`
	}{zoneCreds, url}}
}

func (h *handler) unbind(req *http.Request) responseEntity {
	vars := mux.Vars(req)
	breq := BindingRequest{InstanceId: vars[instanceId], BindingId: vars[bindingId]}

	log.Printf("Handler: Unbinding: %v", breq)

	for _, brokerService := range h.brokerServices {
		if err := brokerService.Unbind(breq); err != nil {
			return handleServiceError(err)
		}
	}

	log.Printf("Handler: Unbound: %v", breq)

	return responseEntity{http.StatusOK, empty}
}

func handleDecodingError(err error) responseEntity {
	log.Printf("Handler: Decoding error: %v", err)
	return responseEntity{http.StatusBadRequest, BrokerError{err.Error()}}
}

func handleServiceError(err error) responseEntity {
	log.Printf("Handler: Service error: %v", err)

	switch err := err.(type) {
	case BrokerServiceError:
		switch err.Code() {
		case ErrCodeConflict:
			return responseEntity{http.StatusConflict, empty}
		case ErrCodeGone:
			return responseEntity{http.StatusGone, empty}
		}
	}
	return responseEntity{http.StatusInternalServerError, BrokerError{err.Error()}}
}
