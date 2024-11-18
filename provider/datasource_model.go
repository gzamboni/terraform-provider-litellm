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
		Description: "",
	},
	"model_info_db_model": {
		Type:        schema.TypeBool,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"model_info_tier": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"model_info_base_model": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"model_info_updated_at": {
		Type:                  schema.TypeString,
		Required:              false,
		Optional:              true,
		Description:           "",
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc:      SupressIgnoreChange,
	},
	"model_info_updated_by": {
		Type:                  schema.TypeString,
		Required:              false,
		Optional:              true,
		Description:           "",
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc:      SupressIgnoreChange,
	},
	"model_info_created_at": {
		Type:                  schema.TypeString,
		Required:              false,
		Optional:              true,
		Description:           "",
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc:      SupressIgnoreChange,
	},
	"model_info_created_by": {
		Type:                  schema.TypeString,
		Required:              false,
		Optional:              true,
		Description:           "",
		DiffSuppressOnRefresh: true,
		DiffSuppressFunc:      SupressIgnoreChange,
	},
}

var DatasourceLitellmParamsSchema = map[string]*schema.Schema{
	"litellm_params_model": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_custom_llm_provider": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_tpm": {
		Type:        schema.TypeInt,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_rpm": {
		Type:        schema.TypeInt,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_api_key": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_api_base": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_api_version": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_timeout": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_stream_timeout": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_max_retries": {
		Type:        schema.TypeInt,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_organization": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_configurable_clientside_auth_params": {
		Type:        schema.TypeList,
		Required:    false,
		Optional:    true,
		Description: "",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"litellm_params_region_name": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_vertex_project": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_vertex_location": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_vertex_credentials": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_aws_access_key_id": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_aws_secret_key_id": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_aws_region_name": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_watsonx_region_name": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_input_cost_per_token": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_output_cost_per_token": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_input_cost_per_second": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_output_cost_per_second": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "",
	},
	"litellm_params_max_file_size_mb": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "",
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
