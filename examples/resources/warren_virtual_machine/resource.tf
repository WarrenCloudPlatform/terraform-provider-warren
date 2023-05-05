resource "warren_virtual_machine" "server42" {
  disk_size_in_gb = 20
  memory          = 1024
  name            = "Server #42"
  username        = "user42"
  os_name         = data.warren_os_base_image.ubuntu.os_name
  os_version      = data.warren_os_base_image.ubuntu.os_version
  vcpu            = 1
}
