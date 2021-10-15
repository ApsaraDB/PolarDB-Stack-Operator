/* 
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */


package controllers

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/configuration"
)

var debugPredicateLogger = ctrl.Log.WithName("predicate").WithName("eventFilters")

type DebugPredicate struct {
	predicate.Funcs
}

func filterOperatorName(runtimeObj runtime.Object) bool {
	if configuration.GetConfig().FilterOperatorName == "" {
		return true
	}
	errMsg := "not for operator " + configuration.GetConfig().FilterOperatorName + " event, do not deal with it"
	if unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(runtimeObj); err != nil {
		debugPredicateLogger.Error(err, "runtime object cannot be convert to unstructured object")
		return false
	} else if specObj, ok := unstructuredObj["spec"]; !ok {
		debugPredicateLogger.Info(errMsg)
		return false
	} else if operator, ok := specObj.(map[string]interface{})["operatorName"]; !ok {
		debugPredicateLogger.Info(errMsg)
		return false
	} else if operator.(string) != configuration.GetConfig().FilterOperatorName {
		debugPredicateLogger.Info(errMsg)
		return false
	}
	return true
}

// Update implements default UpdateEvent filter for validating resource version change
func (DebugPredicate) Update(e event.UpdateEvent) bool {
	return filterOperatorName(e.ObjectNew) && filterOperatorName(e.ObjectOld)
}

// Create implements Predicate
func (p DebugPredicate) Create(e event.CreateEvent) bool {
	return filterOperatorName(e.Object)
}

// Delete implements Predicate
func (p DebugPredicate) Delete(e event.DeleteEvent) bool {
	return filterOperatorName(e.Object)
}

func (p DebugPredicate) Generic(e event.GenericEvent) bool {
	return filterOperatorName(e.Object)
}
