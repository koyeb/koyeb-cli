use assert_cmd::Command;
use predicates::prelude::*;
use std::fs;

#[test]
fn version_prints_package_version() {
    let mut cmd = Command::cargo_bin("koyeb").unwrap();
    cmd.arg("version")
        .assert()
        .success()
        .stdout(predicate::str::contains(env!("CARGO_PKG_VERSION")));
}

#[test]
fn help_mentions_core_commands() {
    let mut cmd = Command::cargo_bin("koyeb").unwrap();
    cmd.arg("--help")
        .assert()
        .success()
        .stdout(predicate::str::contains("apps"))
        .stdout(predicate::str::contains("services"))
        .stdout(predicate::str::contains("sandbox"));
}

#[test]
fn missing_token_is_reported_for_api_commands() {
    let tmp = tempfile::tempdir().unwrap();
    let mut cmd = Command::cargo_bin("koyeb").unwrap();
    cmd.env_remove("KOYEB_TOKEN")
        .env("KOYEB_CONFIG", tmp.path().join("missing.yaml"))
        .args(["apps", "list"])
        .assert()
        .failure()
        .stderr(predicate::str::contains("missing API token"));
}

#[test]
fn login_writes_yaml_config() {
    let tmp = tempfile::tempdir().unwrap();
    let cfg = tmp.path().join("config.yaml");
    let mut cmd = Command::cargo_bin("koyeb").unwrap();
    cmd.arg("--config")
        .arg(&cfg)
        .arg("login")
        .write_stdin("test-token\n")
        .assert()
        .success();
    let content = fs::read_to_string(cfg).unwrap();
    assert!(content.contains("test-token"));
}
