resource "litellm_user" "my_user" {
    user_id = "myuser@mail.com"
    user_alias = "myuser@mail.com"
    user_email = "myuser@mail.com"

    user_role = "proxy_admin"

    send_invite_mail = false

    budget_duration = "30d"
    max_budget = 10.0
    tpm_limit = 2000
    rpm_limit = 10
    models = ["azure/gpt-4o"]

    auto_create_key = true
    metadata = {
        "createdBy": "terraform-provider-litellm"
    }
}