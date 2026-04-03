# Read capacity and utilization for all repositories
data "veeam_repository_states" "all" {}

output "repository_utilization" {
  value = {
    for s in data.veeam_repository_states.all.states :
    s.name => {
      status     = s.status
      free_space = s.free_space
      used_space = s.used_space
      capacity   = s.capacity
    }
  }
}

# Find repositories with less than 10% free space remaining
output "low_space_repos" {
  value = [
    for s in data.veeam_repository_states.all.states : s.name
    if s.capacity > 0 && (s.free_space * 100 / s.capacity) < 10
  ]
}
