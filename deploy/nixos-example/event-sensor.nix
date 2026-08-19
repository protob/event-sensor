let
  user = "username";
  home = "/home/${user}";
  dataDir = "${home}/0/apps-media/event-sensor";
in
{
  version = "0.1.0";
  description = "Event Sensor";
  deployed = true;
  mode = "single-binary";

  api = {
    port = 6034;
    binary = "event-sensor-bin";
    env = {
      portVar = "PORT";       # compiled into the binary (config.go)
      dataVar = "DATA_PATH";  # unused: DB_PATH below is a full file path
    };
  };

  build = {
    frontend_dir = "frontend";
    backend_dir = ".";
    backend_cmd = ".";
    cgo_enabled = false;      # modernc sqlite - static binary, x86_64 + aarch64
  };

  paths = {
    binary = "${home}/custom-systemd/event-sensor/event-sensor-bin";
    data = dataDir;
  };

  # DB_PATH is the database FILE, not a directory - the generator's dataVar
  # injects a directory, so the file path goes through extraEnv instead.
  extraEnv = {
    DB_PATH = "${dataDir}/event-sensor.db";
  };
}
