resource "warren_disk" "disk42" {
  server_uuid = resource.warren_virtual_machine.server42.id
  size_in_gb  = 20
}
