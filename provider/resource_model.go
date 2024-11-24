package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"

	"github.com/gzamboni/terraform-provider-litellm/provider/litellm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SupressIgnoreChange(k, old, new string, d *schema.ResourceData) bool {
	return true
}

var ModelInfoSchema = map[string]*schema.Schema{
	"model_info_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "This attribute is the model identifier.",
	},
	"model_info_db_model": {
		Type:        schema.TypeBool,
		Required:    false,
		Optional:    true,
		Description: "Set to true if the is created through terraform-provider-litellm or through the API. If set to false it will be considered like a config model.",
		Default:     true,
	},
	"model_info_tier": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "Model tier, can be free, enterprise, ...",
	},
	"model_info_base_model": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "From which model is derived this model",
	},
	"model_info_updated_at": {
		Type:                  schema.TypeString,
		Required:              false,
		Optional:              true,
		Description:           "The last time when the model has been updated",
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc:      SupressIgnoreChange,
	},
	"model_info_updated_by": {
		Type:                  schema.TypeString,
		Required:              false,
		Optional:              true,
		Description:           "Who has made the last update to the model",
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc:      SupressIgnoreChange,
	},
	"model_info_created_at": {
		Type:                  schema.TypeString,
		Required:              false,
		Optional:              true,
		Description:           "When the model has been created",
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc:      SupressIgnoreChange,
	},
	"model_info_created_by": {
		Type:                  schema.TypeString,
		Required:              false,
		Optional:              true,
		Description:           "Who created the model",
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc:      SupressIgnoreChange,
	},
}

var LitellmParamsSchema = map[string]*schema.Schema{
	"litellm_params_model": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "To specify which model to use in the api base endpoint",
	},
	"litellm_params_custom_llm_provider": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "To specify a custom provider for an LLM Model, so LITELLM know the provider.",
	},
	"litellm_params_tpm": {
		Type:        schema.TypeInt,
		Required:    false,
		Optional:    true,
		Description: "Max TPM for the model",
	},
	"litellm_params_rpm": {
		Type:        schema.TypeInt,
		Required:    false,
		Optional:    true,
		Description: "Max RPM for the model",
	},
	"litellm_params_api_key": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "Api key to use to authenticate against the api base",
	},
	"litellm_params_api_base": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "Base API Endpoint to use the model",
	},
	"litellm_params_api_version": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "Api version to use when calling the api base",
	},
	"litellm_params_timeout": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "When receiving no response from the llm model, time after which timeout is called",
	},
	"litellm_params_stream_timeout": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "When receiving no response form the llm model, time after which timeout is called",
	},
	"litellm_params_max_retries": {
		Type:        schema.TypeInt,
		Required:    false,
		Optional:    true,
		Description: "Maximum number of retries before returning an error",
	},
	"litellm_params_organization": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "Organization name parameter can be useful for some models",
	},
	"litellm_params_configurable_clientside_auth_params": {
		Type:        schema.TypeList,
		Required:    false,
		Optional:    true,
		Description: "Which params are allowed to modify an litellm user.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"litellm_params_region_name": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "Region name where the llm is located can be useful for some models",
	},
	"litellm_params_vertex_project": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "GCP Project ID (not number) where the vertex service is deployed. Useful for some models",
	},
	"litellm_params_vertex_location": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "GCP Location where the vertex service is deployed. Useful for some models",
	},
	"litellm_params_vertex_credentials": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "GCP credentials to use the vertex LLM Model. Useful for some models",
	},
	"litellm_params_aws_access_key_id": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "AWS Key id, useful for some models",
	},
	"litellm_params_aws_secret_key_id": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "AWS Key secret, useful for some models",
	},
	"litellm_params_aws_region_name": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "AWS Region where the LLM Model is deployed. Useful for some models.",
	},
	"litellm_params_watsonx_region_name": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "Watsonx region name",
	},
	"litellm_params_input_cost_per_token": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "Input cost per token. How much cost a single input token not per 1k tokens or 1M tokens.",
	},
	"litellm_params_output_cost_per_token": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "Output cost per token. How much cost a single output token not per 1k tokens or 1M tokens.",
	},
	"litellm_params_input_cost_per_second": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "Input cost per second.",
	},
	"litellm_params_output_cost_per_second": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "Output cost per second.",
	},
	"litellm_params_max_file_size_mb": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "Max file size allowed to upload",
	},
}

var BaseModelSchema = map[string]*schema.Schema{
	"model_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the model to be managed.",
	},
}

func ModelSchema() map[string]*schema.Schema {
	ModelSchema := make(map[string]*schema.Schema)
	maps.Copy(ModelSchema, BaseModelSchema)
	maps.Copy(ModelSchema, LitellmParamsSchema)
	maps.Copy(ModelSchema, ModelInfoSchema)
	return ModelSchema
}

func getModelInfoFromResourceData(d *schema.ResourceData) litellm.ModelInfo {
	data := litellm.ModelInfo{
		Id:        d.Get("model_info_id").(string),
		DbModel:   d.Get("model_info_db_model").(bool),
		Tier:      d.Get("model_info_tier").(string),
		BaseModel: d.Get("model_info_base_model").(string),
		UpdatedAt: d.Get("model_info_updated_at").(string),
		UpdatedBy: d.Get("model_info_updated_by").(string),
		CreatedAt: d.Get("model_info_created_at").(string),
		CreatedBy: d.Get("model_info_created_by").(string),
	}

	return data
}

func setModelInfoFromModel(model *litellm.Model, d *schema.ResourceData) {
	d.Set("model_info_id", model.ModelInfo.Id)
	d.Set("model_info_db_model", model.ModelInfo.DbModel)
	d.Set("model_info_tier", model.ModelInfo.Tier)
	d.Set("model_info_base_model", model.ModelInfo.BaseModel)
	d.Set("model_info_updated_at", model.ModelInfo.UpdatedAt)
	d.Set("model_info_updated_by", model.ModelInfo.UpdatedBy)
	d.Set("model_info_created_at", model.ModelInfo.CreatedAt)
	d.Set("model_info_created_by", model.ModelInfo.CreatedBy)
}

func getLitellmParamsFromResourceData(d *schema.ResourceData) litellm.LitellmParams {
	rawConfigurableClientsideConfigParams := d.Get("litellm_params_configurable_clientside_auth_params").([]interface{})
	var configurableClientsideConfigParams []string
	for _, item := range rawConfigurableClientsideConfigParams {
		configurableClientsideConfigParams = append(configurableClientsideConfigParams, item.(string))
	}

	data := litellm.LitellmParams{
		Model:                            d.Get("litellm_params_model").(string),
		CustomLlmProvider:                d.Get("litellm_params_custom_llm_provider").(string),
		Tpm:                              d.Get("litellm_params_tpm").(int),
		Rpm:                              d.Get("litellm_params_rpm").(int),
		ApiKey:                           d.Get("litellm_params_api_key").(string),
		ApiBase:                          d.Get("litellm_params_api_base").(string),
		ApiVersion:                       d.Get("litellm_params_api_version").(string),
		Timeout:                          d.Get("litellm_params_timeout").(float64),
		StreamTimeout:                    d.Get("litellm_params_stream_timeout").(float64),
		MaxRetries:                       d.Get("litellm_params_max_retries").(int),
		Organization:                     d.Get("litellm_params_organization").(string),
		ConfigurableClientsideAuthParams: configurableClientsideConfigParams,
		RegionName:                       d.Get("litellm_params_region_name").(string),
		VertexProject:                    d.Get("litellm_params_vertex_project").(string),
		VertexLocation:                   d.Get("litellm_params_vertex_location").(string),
		VertexCredentials:                d.Get("litellm_params_vertex_credentials").(string),
		AwsAccessKeyId:                   d.Get("litellm_params_aws_access_key_id").(string),
		AwsSecretKeyId:                   d.Get("litellm_params_aws_secret_key_id").(string),
		AwsRegionName:                    d.Get("litellm_params_aws_region_name").(string),
		WatsonxRegionName:                d.Get("litellm_params_watsonx_region_name").(string),
		InputCostPerToken:                d.Get("litellm_params_input_cost_per_token").(float64),
		OutputCostPerToken:               d.Get("litellm_params_output_cost_per_token").(float64),
		OutputCostPerSecond:              d.Get("litellm_params_output_cost_per_second").(float64),
		InputCostPerSecond:               d.Get("litellm_params_input_cost_per_second").(float64),
		MaxFileSizeMB:                    d.Get("litellm_params_max_file_size_mb").(float64),
	}

	return data
}

func setLitellmParamsFromModel(model *litellm.Model, d *schema.ResourceData) {
	d.Set("litellm_params_model", model.LitellmParams.Model)
	d.Set("litellm_params_custom_llm_provider", model.LitellmParams.CustomLlmProvider)
	d.Set("litellm_params_tpm", model.LitellmParams.Tpm)
	d.Set("litellm_params_rpm", model.LitellmParams.Rpm)
	d.Set("litellm_params_api_key", model.LitellmParams.ApiKey)
	d.Set("litellm_params_api_base", model.LitellmParams.ApiBase)
	d.Set("litellm_params_api_version", model.LitellmParams.ApiVersion)
	d.Set("litellm_params_timeout", model.LitellmParams.Timeout)
	d.Set("litellm_params_stream_timeout", model.LitellmParams.StreamTimeout)
	d.Set("litellm_params_max_retries", model.LitellmParams.MaxRetries)
	d.Set("litellm_params_organization", model.LitellmParams.Organization)
	d.Set("litellm_params_configurable_clientside_auth_params", model.LitellmParams.ConfigurableClientsideAuthParams)
	d.Set("litellm_params_region_name", model.LitellmParams.RegionName)
	d.Set("litellm_params_vertex_project", model.LitellmParams.VertexProject)
	d.Set("litellm_params_vertex_location", model.LitellmParams.VertexLocation)
	d.Set("litellm_params_vertex_credentials", model.LitellmParams.VertexCredentials)
	d.Set("litellm_params_aws_access_key_id", model.LitellmParams.AwsAccessKeyId)
	d.Set("litellm_params_aws_secret_key_id", model.LitellmParams.AwsSecretKeyId)
	d.Set("litellm_params_aws_region_name", model.LitellmParams.AwsRegionName)
	d.Set("litellm_params_watsonx_region_name", model.LitellmParams.WatsonxRegionName)
	d.Set("litellm_params_input_cost_per_token", model.LitellmParams.InputCostPerToken)
	d.Set("litellm_params_input_cost_per_second", model.LitellmParams.InputCostPerSecond)
	d.Set("litellm_params_output_cost_per_token", model.LitellmParams.OutputCostPerToken)
	d.Set("litellm_params_output_cost_per_second", model.LitellmParams.OutputCostPerSecond)
	d.Set("litellm_params_max_file_size_mb", model.LitellmParams.MaxFileSizeMB)
}

func getModelFromResourceData(d *schema.ResourceData) litellm.Model {
	data := litellm.Model{
		ModelName:     d.Get("model_name").(string),
		LitellmParams: getLitellmParamsFromResourceData(d),
		ModelInfo:     getModelInfoFromResourceData(d),
	}

	return data
}

func setResourceDataFromModel(model *litellm.Model, d *schema.ResourceData) {
	setLitellmParamsFromModel(model, d)
	setModelInfoFromModel(model, d)
	d.Set("model_name", model.ModelName)
}

func resourceModel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceModelCreate,
		ReadContext:   resourceModelRead,
		UpdateContext: resourceModelUpdate,
		DeleteContext: resourceModelDelete,
		Schema:        ModelSchema(),
		Description: `The ` + "`litellm_model`" + ` resource allows you to manage AI language models within your LiteLLM Proxy instance using Terraform.
This resource enables you to automate the creation, updating, and deletion of models, integrating AI model lifecycle
management into your infrastructure as code practices.

By using ` + "`litellm_model`" + `, you can seamlessly incorporate AI capabilities into your applications and services, while
maintaining consistency and version control through Terraform's declarative configurations.`,
	}
}

func resourceModelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	requestBody := getModelFromResourceData(d)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/model/new", client.ApiBaseURL)
	req, err := client.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	d.SetId(requestBody.ModelInfo.Id)

	return diags
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*LitellmClient)

	stateModel := getModelFromResourceData(d)

	params := url.Values{}
	params.Add("litellm_model_id", stateModel.ModelInfo.Id)
	apiUrl := fmt.Sprintf("%s/model/info?%s", client.ApiBaseURL, params.Encode())

	req, err := client.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		diag.Errorf("API Request to read model info returned status code : %d", resp.StatusCode)
	}

	var jsonPayload litellm.ModelInfoWModelId
	err = json.NewDecoder(resp.Body).Decode(&jsonPayload)
	if err != nil {
		diag.FromErr(err)
	}
	rModel := &jsonPayload.Data

	setResourceDataFromModel(rModel, d)

	return diags
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	requestBody := getModelFromResourceData(d)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/model/update", client.ApiBaseURL)
	req, err := client.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	return diags
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	model := getModelFromResourceData(d)

	jsonPayload := map[string]interface{}{
		"id": model.ModelInfo.Id,
	}

	jsonData, err := json.Marshal(jsonPayload)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/model/delete", client.ApiBaseURL)
	req, err := client.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	d.SetId("")

	return diags
}
