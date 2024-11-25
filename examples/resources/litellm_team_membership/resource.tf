resource "litellm_team_membership" "example_user_team_association" {
  user_id = litellm_user.example_user.user_id
  team_id = litellm_team.example_team.team_id
  role    = "internal_user_viewer"
}