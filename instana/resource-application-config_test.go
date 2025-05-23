package instana_test

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"

	. "github.com/gessnerfl/terraform-provider-instana/instana"
	"github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	"github.com/gessnerfl/terraform-provider-instana/testutils"
)

const resourceApplicationConfigWithTagFilterDefinitionTemplate = `
resource "instana_application_config" "example" {
  label = "name %d"
  scope = "INCLUDE_ALL_DOWNSTREAM"
  boundary_scope = "ALL"
  tag_filter = "%s"
}
`

// Important if a match specification is not provided only the tag filter will be available.
const serverResponseWithTagFilterTemplate = `
{
	"id" : "%s",
	"label" : "name %d",
	"scope" : "INCLUDE_ALL_DOWNSTREAM",
	"boundaryScope" : "ALL",
	"tagFilterExpression" : {
		"type" : "EXPRESSION",
		"logicalOperator": "OR",
		"elements" : [
			{
				"type" : "EXPRESSION",
				"logicalOperator": "AND",
				"elements" : [
					{
						"type" : "TAG_FILTER",
						"name" : "entity.name",
						"entity" : "DESTINATION",
						"operator" : "CONTAINS",
						"stringValue" : "foo",
						"value" : "foo"
					},
					{
						"type" : "TAG_FILTER",
						"name" : "agent.tag",
						"entity" : "DESTINATION",
						"operator" : "EQUALS",
						"stringValue" : "environment=dev-local-gessnerfl",
						"key": "environment",
						"value": "dev-local-gessnerfl"
					}
				]
			},
			{
				"type" : "TAG_FILTER",
				"name" : "call.http.status",
				"entity" : "NOT_APPLICABLE",
				"operator" : "EQUALS",
				"numberValue" : 404,
				"value" : 404
			}
		]
	}
}
`

const (
	testApplicationConfigDefinition = "instana_application_config.example"
	defaultTagFilter                = "entity.name CONTAINS 'foo' AND agent.tag:environment EQUALS 'dev-local-gessnerfl' OR call.http.status@na EQUALS 404"
	defaultNormalizedTagFilter      = "((entity.name@dest CONTAINS 'foo' AND agent.tag:'environment'@dest EQUALS 'dev-local-gessnerfl') OR call.http.status@na EQUALS 404)"
	defaultLabel                    = "label"
	entityName                      = "entity.name"
	expressionEntityTypeDestEqValue = "entity.type@dest EQUALS 'foo'"
	expressionEntityTypeSrcEqValue  = "entity.type@src EQUALS 'foo'"
)

var defaultTagFilterModel = restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{
	restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{
		restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, entityName, restapi.ContainsOperator, "foo"),
		restapi.NewTagTagFilter(restapi.TagFilterEntityDestination, "agent.tag", restapi.EqualsOperator, "environment", "dev-local-gessnerfl"),
	}),
	restapi.NewNumberTagFilter(restapi.TagFilterEntityNotApplicable, "call.http.status", restapi.EqualsOperator, 404),
})

const applicationConfigID = "application-config-id"

func TestCRUDOfApplicationConfigWithTagFilterResourceWithMockServer(t *testing.T) {
	httpServer := createMockHttpServerForApplicationConfig(restapi.ApplicationConfigsResourcePath, serverResponseWithTagFilterTemplate)
	httpServer.Start()
	defer httpServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testProviderFactory,
		Steps: []resource.TestStep{
			createApplicationConfigWithTagFilterResourceTestStep(httpServer.GetPort(), 0),
			testStepImport(testApplicationConfigDefinition),
			createApplicationConfigWithTagFilterResourceTestStep(httpServer.GetPort(), 1),
			testStepImport(testApplicationConfigDefinition),
		},
	})
}

func createApplicationConfigWithTagFilterResourceTestStep(httpPort int, iteration int) resource.TestStep {
	config := appendProviderConfig(fmt.Sprintf(resourceApplicationConfigWithTagFilterDefinitionTemplate, iteration, defaultTagFilter), httpPort)
	return resource.TestStep{
		Config: config,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(testApplicationConfigDefinition, "id"),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldLabel, formatResourceName(iteration)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeAllDownstream)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldTagFilter, defaultNormalizedTagFilter),
		),
	}
}

func createMockHttpServerForApplicationConfig(resourcePath string, responseTemplate string, templateVars ...interface{}) testutils.TestHTTPServer {
	pathTemplate := resourcePath + "/{id}"
	httpServer := testutils.NewTestHTTPServer()
	//handles getting, deleting and updating existing application configs
	responseHandler := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		path := resourcePath + "/" + vars["id"]
		callCount := httpServer.GetCallCount(http.MethodPut, path)
		var json string
		if templateVars != nil {
			json = formatResponseTemplate(responseTemplate, vars["id"], callCount, templateVars...)
		} else {
			json = formatResponseTemplate(responseTemplate, vars["id"], callCount)
		}
		w.Header().Set(contentType, r.Header.Get(contentType))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(json))
		if err != nil {
			log.Fatalf("failed to write response: %s", err)
		}
	}

	//handles creating applications configs
	postHandler := func(w http.ResponseWriter, r *http.Request) {
		callCount := getZeroBasedCallCount(httpServer, http.MethodPost, resourcePath)
		var json string
		if templateVars != nil {
			json = formatResponseTemplate(responseTemplate, "UdkQGmswQ8W05YSdYqysNQ", callCount, templateVars...)
		} else {
			json = formatResponseTemplate(responseTemplate, "UdkQGmswQ8W05YSdYqysNQ", callCount)
		}
		w.Header().Set(contentType, r.Header.Get(contentType))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(json))
		if err != nil {
			log.Fatalf("failed to write response: %s", err)
		}
	}
	httpServer.AddRoute(http.MethodPut, pathTemplate, responseHandler)
	httpServer.AddRoute(http.MethodPost, resourcePath, postHandler)
	httpServer.AddRoute(http.MethodDelete, pathTemplate, responseHandler)
	httpServer.AddRoute(http.MethodGet, pathTemplate, responseHandler)
	return httpServer
}

func TestApplicationConfigSchemaDefinitionIsValid(t *testing.T) {
	schema := NewApplicationConfigResourceHandle().MetaData().Schema

	schemaAssert := testutils.NewTerraformSchemaAssert(schema, t)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(ApplicationConfigFieldLabel)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeStringWithDefault(ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeNoDownstream))
	schemaAssert.AssertSchemaIsOptionalAndOfTypeStringWithDefault(ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeDefault))
	schemaAssert.AssertSchemaIsOptionalAndOfTypeString(ApplicationConfigFieldTagFilter)
}

func TestApplicationConfigResourceShouldHaveSchemaVersionFour(t *testing.T) {
	require.Equal(t, 4, NewApplicationConfigResourceHandle().MetaData().SchemaVersion)
}

func TestApplicationConfigResourceShouldHaveFourStateUpgraderForVersionZeroToThree(t *testing.T) {
	resourceHandler := NewApplicationConfigResourceHandle()

	require.Equal(t, 1, len(resourceHandler.StateUpgraders()))
	require.Equal(t, 3, resourceHandler.StateUpgraders()[0].Version)
}

func TestApplicationConfigShouldMigrateFullLabelToLabelWhenExecutingV3StateUpgraderAndFullLabelIsAvailable(t *testing.T) {
	input := map[string]interface{}{
		"full_label": "test",
	}
	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[0].Upgrade(nil, input, nil)

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.NotContains(t, result, ApplicationConfigFieldFullLabel)
	require.Contains(t, result, ApplicationConfigFieldLabel)
	require.Equal(t, "test", result[ApplicationConfigFieldLabel])
}

func TestApplicationConfigShouldDeleteMatchSpecificationWhenExecutingV3StateUpgraderAndMatchSpecificationIsAvailable(t *testing.T) {
	input := map[string]interface{}{
		"match_specification": "test",
	}
	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[0].Upgrade(nil, input, nil)

	require.NoError(t, err)
	require.Len(t, result, 0)
	require.NotContains(t, result, "match_specification")
}

func TestApplicationConfigShouldDoNothingWhenExecutingV3StateUpgraderAndFullLabelandMatchSpecificationIsNotAvailable(t *testing.T) {
	input := map[string]interface{}{
		"label": "test",
	}
	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[0].Upgrade(nil, input, nil)

	require.NoError(t, err)
	require.Equal(t, input, result)
}

func TestShouldReturnCorrectResourceNameForApplicationConfigResource(t *testing.T) {
	name := NewApplicationConfigResourceHandle().MetaData().ResourceName

	require.Equal(t, name, "instana_application_config")
}

func TestShouldUpdateApplicationConfigTerraformResourceStateFromModelWhenTagFilterIsProvided(t *testing.T) {
	applicationConfig := restapi.ApplicationConfig{
		ID:                  applicationConfigID,
		Label:               defaultLabel,
		TagFilterExpression: defaultTagFilterModel,
		Scope:               restapi.ApplicationConfigScopeIncludeNoDownstream,
		BoundaryScope:       restapi.BoundaryScopeAll,
	}

	testHelper := NewTestHelper[*restapi.ApplicationConfig](t)
	sut := NewApplicationConfigResourceHandle()
	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, &applicationConfig)

	require.NoError(t, err)
	require.Equal(t, applicationConfigID, resourceData.Id())
	require.Equal(t, defaultLabel, resourceData.Get(ApplicationConfigFieldLabel))
	require.Equal(t, defaultNormalizedTagFilter, resourceData.Get(ApplicationConfigFieldTagFilter))
	require.Equal(t, string(restapi.ApplicationConfigScopeIncludeNoDownstream), resourceData.Get(ApplicationConfigFieldScope))
	require.Equal(t, string(restapi.BoundaryScopeAll), resourceData.Get(ApplicationConfigFieldBoundaryScope))
}

func TestShouldFailToUpdateApplicationConfigTerraformResourceStateFromModelWhenTagFilterIsNotValid(t *testing.T) {
	comparison := restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, entityName, "INVALID", "foo")
	applicationConfig := restapi.ApplicationConfig{
		ID:                  applicationConfigID,
		Label:               defaultLabel,
		TagFilterExpression: comparison,
		Scope:               restapi.ApplicationConfigScopeIncludeNoDownstream,
	}

	testHelper := NewTestHelper[*restapi.ApplicationConfig](t)
	sut := NewApplicationConfigResourceHandle()
	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, &applicationConfig)

	require.Error(t, err)
}

func TestShouldSuccessfullyConvertApplicationConfigStateToDataModelWhenTagFilterIsAvailable(t *testing.T) {
	testHelper := NewTestHelper[*restapi.ApplicationConfig](t)
	resourceHandle := NewApplicationConfigResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)
	resourceData.SetId(applicationConfigID)
	setValueOnResourceData(t, resourceData, ApplicationConfigFieldLabel, defaultLabel)
	setValueOnResourceData(t, resourceData, ApplicationConfigFieldTagFilter, defaultTagFilter)
	setValueOnResourceData(t, resourceData, ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeNoDownstream))
	setValueOnResourceData(t, resourceData, ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll))

	result, err := resourceHandle.MapStateToDataObject(resourceData)

	require.Nil(t, err)
	require.IsType(t, &restapi.ApplicationConfig{}, result)
	require.Equal(t, applicationConfigID, result.GetIDForResourcePath())
	require.Equal(t, defaultLabel, result.Label)
	require.Equal(t, defaultTagFilterModel, result.TagFilterExpression)
	require.Equal(t, restapi.ApplicationConfigScopeIncludeNoDownstream, result.Scope)
	require.Equal(t, restapi.BoundaryScopeAll, result.BoundaryScope)
}

func TestShouldFailToConvertApplicationConfigStateToDataModelWhenTagFilterIsNotValid(t *testing.T) {
	testHelper := NewTestHelper[*restapi.ApplicationConfig](t)
	resourceHandle := NewApplicationConfigResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)
	resourceData.SetId(applicationConfigID)
	setValueOnResourceData(t, resourceData, ApplicationConfigFieldLabel, defaultLabel)
	setValueOnResourceData(t, resourceData, ApplicationConfigFieldTagFilter, "INVALID")
	setValueOnResourceData(t, resourceData, ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeNoDownstream))
	setValueOnResourceData(t, resourceData, ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll))

	_, err := resourceHandle.MapStateToDataObject(resourceData)

	require.NotNil(t, err)
}
