package provider

import (
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

var DatasourceModelInfoSchema = map[string]*schema.Schema{
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

var DatasourceLitellmParamsSchema = map[string]*schema.Schema{
	"litellm_params_model": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
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

var DatasourceBaseModelSchema = map[string]*schema.Schema{
	"model_name": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "Name of the model to be managed.",
	},
}

func DatasourceModelSchema() map[string]*schema.Schema {
	ModelSchema := make(map[string]*schema.Schema)
	maps.Copy(ModelSchema, DatasourceBaseModelSchema)
	maps.Copy(ModelSchema, DatasourceLitellmParamsSchema)
	maps.Copy(ModelSchema, DatasourceModelInfoSchema)
	return ModelSchema
}

func datasourceModel() *schema.Resource {
	return &schema.Resource{
		ReadContext: DatasourceResourceModelRead,
		Schema:      DatasourceModelSchema(),
		Description: `This data source allow you to fetch data from an existing model in your litellm instance.
		
		For a config model be careful the ` + "`model_info_id`" + ` is an UUID it's not the name like ` + "`azure/gpt-4o`" + `.
		`,
	}
}

func DatasourceResourceModelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*LitellmClient)

	modelId := d.Get("model_info_id").(string)

	params := url.Values{}
	params.Add("litellm_model_id", modelId)
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
		diag.Errorf("API Request to read model info returned status code %d", resp.StatusCode)
	}

	var jsonPayload litellm.ModelInfoWModelId
	err = json.NewDecoder(resp.Body).Decode(&jsonPayload)
	if err != nil {
		diag.FromErr(err)
	}
	rModel := &jsonPayload.Data

	d.SetId(rModel.ModelInfo.Id)
	setResourceDataFromModel(rModel, d)

	return diags
}
