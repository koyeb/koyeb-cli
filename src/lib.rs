use anyhow::{Context, Result, anyhow, bail};
use clap::{ArgAction, Args, CommandFactory, Parser, Subcommand, ValueEnum};

use serde_json::{Value, json};
use std::collections::BTreeMap;
use std::fmt;
use std::fs;
use std::io::{self, Read, Write};
use std::path::{Path, PathBuf};
use std::time::Duration;

pub const DEFAULT_API_URL: &str = "https://app.koyeb.com";
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Parser, Debug)]
#[command(name = "koyeb", about = "Koyeb CLI", disable_version_flag = true)]
pub struct Cli {
    #[arg(
        short,
        long,
        global = true,
        help = "config file (default is $HOME/.koyeb.yaml, or $KOYEB_CONFIG if set)"
    )]
    config: Option<PathBuf>,
    #[arg(short = 'o', long, global = true, value_enum, default_value_t = OutputFormat::Table, help = "output format (yaml,json,table)")]
    output: OutputFormat,
    #[arg(short = 'd', long, global = true, action = ArgAction::SetTrue, help = "enable the debug output")]
    debug: bool,
    #[arg(long, global = true, action = ArgAction::SetTrue, help = "do not hide sensitive information (tokens) in the debug output")]
    debug_full: bool,
    #[arg(long, global = true, action = ArgAction::SetTrue, help = "only output ascii characters (no unicode emojis)")]
    _force_ascii: bool,
    #[arg(long, global = true, action = ArgAction::SetTrue, help = "do not truncate output")]
    full: bool,
    #[arg(long, global = true, default_value = DEFAULT_API_URL, help = "url of the api")]
    url: String,
    #[arg(long, global = true, env = "KOYEB_TOKEN", help = "API token")]
    token: Option<String>,
    #[arg(
        long,
        global = true,
        env = "KOYEB_ORGANIZATION",
        help = "organization ID"
    )]
    organization: Option<String>,
    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Clone, Copy, Debug, ValueEnum, PartialEq, Eq)]
pub enum OutputFormat {
    Table,
    Json,
    Yaml,
}

impl fmt::Display for OutputFormat {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(match self {
            OutputFormat::Table => "table",
            OutputFormat::Json => "json",
            OutputFormat::Yaml => "yaml",
        })
    }
}

#[derive(Subcommand, Debug)]
enum Commands {
    #[command(about = "Login to your Koyeb account")]
    Login,
    #[command(about = "Get version")]
    Version,
    #[command(about = "Generate completion script")]
    Completion { shell: CompletionShell },
    #[command(visible_aliases = ["a", "app"], about = "Apps")]
    Apps {
        #[command(subcommand)]
        command: AppsCommand,
    },
    #[command(visible_alias = "archive", about = "Archives")]
    Archives {
        #[command(subcommand)]
        command: ArchivesCommand,
    },
    #[command(about = "Deploy a directory to Koyeb")]
    Deploy(DeployArgs),
    #[command(visible_aliases = ["dom", "domain"], about = "Domains")]
    Domains {
        #[command(subcommand)]
        command: DomainsCommand,
    },
    #[command(visible_aliases = ["organization", "orgas", "orga", "orgs", "org", "organisations", "organisation"], about = "Organization")]
    Organizations {
        #[command(subcommand)]
        command: OrganizationsCommand,
    },
    #[command(visible_aliases = ["sec", "secret"], about = "Secrets")]
    Secrets {
        #[command(subcommand)]
        command: SecretsCommand,
    },
    #[command(visible_aliases = ["s", "svc", "service"], about = "Services")]
    Services {
        #[command(subcommand)]
        command: ServicesCommand,
    },
    #[command(about = "Deployments")]
    Deployments {
        #[command(subcommand)]
        command: DeploymentsCommand,
    },
    #[command(visible_aliases = ["rd", "rdep", "rdepl", "rdeploy", "rdeployment", "regional-deployment"], about = "Regional deployments")]
    RegionalDeployments {
        #[command(subcommand)]
        command: RegionalDeploymentsCommand,
    },
    #[command(about = "Instances")]
    Instances {
        #[command(subcommand)]
        command: InstancesCommand,
    },
    #[command(visible_aliases = ["db", "database"], about = "Databases")]
    Databases {
        #[command(subcommand)]
        command: DatabasesCommand,
    },
    #[command(visible_alias = "metric", about = "Metrics")]
    Metrics {
        #[command(subcommand)]
        command: MetricsCommand,
    },
    #[command(visible_aliases = ["vol", "volume"], about = "Manage persistent volumes")]
    Volumes {
        #[command(subcommand)]
        command: VolumesCommand,
    },
    #[command(visible_alias = "snapshot", about = "Manage snapshots")]
    Snapshots {
        #[command(subcommand)]
        command: SnapshotsCommand,
    },
    #[command(about = "Create Koyeb resources from a koyeb-compose.yaml file")]
    Compose(ComposeArgs),
    #[command(
        visible_alias = "sb",
        about = "Sandbox - interactive execution environments"
    )]
    Sandbox {
        #[command(subcommand)]
        command: SandboxCommand,
    },
    #[command(about = "Show information about the currently authenticated user or organization")]
    Whoami,
}

#[derive(ValueEnum, Clone, Debug)]
enum CompletionShell {
    Bash,
    Zsh,
    Fish,
    Powershell,
}

#[derive(Debug, Args)]
struct DeployArgs {
    path: PathBuf,
    target: String,
    #[arg(long)]
    app: Option<String>,
    #[arg(long)]
    wait: bool,
    #[arg(long, default_value = "5m")]
    wait_timeout: HumanDuration,
    #[command(flatten)]
    service: ServiceDefinitionArgs,
}

#[derive(Subcommand, Debug)]
enum AppsCommand {
    #[command(about = "Create app")]
    Create {
        name: String,
        #[arg(long)]
        delete_when_empty: bool,
    },
    #[command(about = "Create app and service")]
    Init(InitAppArgs),
    #[command(about = "Get app")]
    Get { name: String },
    #[command(about = "List apps")]
    List,
    #[command(about = "Describe app")]
    Describe { name: String },
    #[command(about = "Update app")]
    Update {
        name: String,
        #[arg(short, long)]
        name_new: Option<String>,
        #[arg(short = 'D', long)]
        domain: Option<String>,
        #[arg(long)]
        delete_when_empty: bool,
    },
    #[command(about = "Delete app")]
    Delete { name: String },
    #[command(about = "Pause app")]
    Pause { name: String },
    #[command(about = "Resume app")]
    Resume { name: String },
}

#[derive(Debug, Args)]
struct InitAppArgs {
    name: String,
    #[arg(long)]
    wait: bool,
    #[arg(long, default_value = "5m")]
    wait_timeout: HumanDuration,
    #[command(flatten)]
    service: ServiceDefinitionArgs,
}

#[derive(Subcommand, Debug)]
enum ArchivesCommand {
    #[command(about = "Create archive")]
    Create {
        name: String,
        #[arg(long = "ignore-dir", default_values_t = vec![".git".to_string(), "node_modules".to_string(), "vendor".to_string()])]
        ignore_dir: Vec<String>,
    },
}

#[derive(Subcommand, Debug)]
enum DomainsCommand {
    #[command(about = "Get domain")]
    Get { name: String },
    #[command(about = "Create domain")]
    Create {
        name: String,
        #[arg(long = "attach-to")]
        attach_to: Option<String>,
    },
    #[command(about = "Describe domain")]
    Describe { name: String },
    #[command(about = "List domains")]
    List,
    #[command(about = "Delete domain")]
    Delete { name: String },
    #[command(about = "Refresh a custom domain verification status")]
    Refresh { name: String },
    #[command(about = "Attach a custom domain to an existing app")]
    Attach { name: String, app: String },
    #[command(about = "Detach a custom domain from the app it is currently attached to")]
    Detach { name: String },
}

#[derive(Subcommand, Debug)]
enum OrganizationsCommand {
    #[command(about = "List organizations")]
    List,
    #[command(about = "Switch the CLI context to another organization")]
    Switch { organization: String },
}

#[derive(Subcommand, Debug)]
enum SecretsCommand {
    #[command(about = "Create secret")]
    Create {
        name: String,
        #[arg(long = "type", value_enum, default_value_t = SecretKind::Simple)]
        kind: SecretKind,
        #[arg(long)]
        value: Option<String>,
        #[arg(long)]
        filename: Option<PathBuf>,
        #[arg(long = "registry-server")]
        registry_server: Option<String>,
        #[arg(long = "registry-username")]
        registry_username: Option<String>,
        #[arg(long = "registry-password")]
        registry_password: Option<String>,
    },
    #[command(about = "Get secret")]
    Get { name: String },
    #[command(about = "List secrets")]
    List,
    #[command(about = "Describe secret")]
    Describe { name: String },
    #[command(about = "Update secret")]
    Update {
        name: String,
        #[arg(long)]
        value: Option<String>,
        #[arg(long)]
        filename: Option<PathBuf>,
    },
    #[command(about = "Delete secret")]
    Delete { name: String },
    #[command(visible_alias = "show", about = "Show secret value")]
    Reveal { name: String },
}

#[derive(ValueEnum, Clone, Copy, Debug)]
enum SecretKind {
    Simple,
    Registry,
}

#[derive(Subcommand, Debug)]
enum ServicesCommand {
    #[command(about = "Create service")]
    Create(ServiceCreateArgs),
    #[command(about = "Get service")]
    Get {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
    },
    #[command(
        about = "Show unapplied changes saved with the --save-only flag, which will be applied in the next deployment"
    )]
    UnappliedChanges {
        service_name: String,
        #[arg(short, long)]
        app: Option<String>,
    },
    #[command(visible_aliases = ["l", "log"], about = "Get the service logs")]
    Logs(LogsArgs),
    #[command(about = "List services")]
    List {
        #[arg(short, long)]
        app: Option<String>,
        #[arg(short, long)]
        name: Option<String>,
    },
    #[command(about = "Describe service")]
    Describe {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
    },
    #[command(visible_aliases = ["run", "attach"], about = "Run a command in the context of an instance selected among the service instances", trailing_var_arg = true)]
    Exec {
        name: String,
        cmd: String,
        args: Vec<String>,
        #[arg(short, long)]
        app: Option<String>,
    },
    #[command(about = "Update service")]
    Update(ServiceUpdateArgs),
    #[command(about = "Redeploy service")]
    Redeploy {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
        #[arg(long)]
        skip_build: bool,
        #[arg(long)]
        wait: bool,
        #[arg(long, default_value = "5m")]
        wait_timeout: HumanDuration,
        #[arg(long)]
        use_cache: bool,
    },
    #[command(about = "Delete service")]
    Delete {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
    },
    #[command(about = "Pause service")]
    Pause {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
    },
    #[command(about = "Resume service")]
    Resume {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
    },
    #[command(
        about = "Set manual scaling configuration for service (replaces existing configuration)"
    )]
    Scale(ScaleArgs),
}

#[derive(Debug, Args)]
struct ServiceCreateArgs {
    name: String,
    #[arg(short, long)]
    app: Option<String>,
    #[arg(long)]
    wait: bool,
    #[arg(long, default_value = "5m")]
    wait_timeout: HumanDuration,
    #[command(flatten)]
    definition: ServiceDefinitionArgs,
}

#[derive(Debug, Args)]
struct ServiceUpdateArgs {
    name: String,
    #[arg(short, long)]
    app: Option<String>,
    #[arg(long = "name")]
    new_name: Option<String>,
    #[arg(long)]
    override_: bool,
    #[arg(long)]
    skip_build: bool,
    #[arg(long)]
    save_only: bool,
    #[arg(long)]
    wait: bool,
    #[arg(long, default_value = "5m")]
    wait_timeout: HumanDuration,
    #[command(flatten)]
    definition: ServiceDefinitionArgs,
}

#[derive(Debug, Args, Default, Clone)]
struct ServiceDefinitionArgs {
    #[arg(long)]
    git: Option<String>,
    #[arg(long = "git-branch")]
    git_branch: Option<String>,
    #[arg(long = "git-sha")]
    git_sha: Option<String>,
    #[arg(long = "git-builder")]
    git_builder: Option<String>,
    #[arg(long = "git-build-command")]
    git_build_command: Option<String>,
    #[arg(long = "git-run-command")]
    git_run_command: Option<String>,
    #[arg(long = "git-privileged")]
    git_privileged: bool,
    #[arg(long = "git-docker-dockerfile")]
    git_docker_dockerfile: Option<String>,
    #[arg(long = "git-docker-entrypoint")]
    git_docker_entrypoint: Vec<String>,
    #[arg(long = "git-docker-args")]
    git_docker_args: Vec<String>,
    #[arg(long)]
    docker: Option<String>,
    #[arg(long = "docker-entrypoint")]
    docker_entrypoint: Vec<String>,
    #[arg(long = "docker-args")]
    docker_args: Vec<String>,
    #[arg(long)]
    archive: Option<String>,
    #[arg(long = "archive-builder")]
    archive_builder: Option<String>,
    #[arg(long = "archive-build-command")]
    archive_build_command: Option<String>,
    #[arg(long = "archive-run-command")]
    archive_run_command: Option<String>,
    #[arg(long = "archive-docker-dockerfile")]
    archive_docker_dockerfile: Option<String>,
    #[arg(long = "archive-docker-entrypoint")]
    archive_docker_entrypoint: Vec<String>,
    #[arg(long = "archive-docker-args")]
    archive_docker_args: Vec<String>,
    #[arg(long = "archive-ignore-dir", default_values_t = vec![".git".to_string(), "node_modules".to_string(), "vendor".to_string()])]
    archive_ignore_dir: Vec<String>,
    #[arg(long = "instance-type")]
    instance_type: Option<String>,
    #[arg(long = "regions")]
    regions: Vec<String>,
    #[arg(long = "ports")]
    ports: Vec<String>,
    #[arg(long = "routes")]
    routes: Vec<String>,
    #[arg(long = "env")]
    env: Vec<String>,
    #[arg(long = "env-file")]
    env_file: Vec<PathBuf>,
    #[arg(long = "secret")]
    secrets: Vec<String>,
    #[arg(long = "config-file")]
    config_file: Vec<String>,
    #[arg(long = "volumes")]
    volumes: Vec<String>,
    #[arg(long = "checks")]
    checks: Vec<String>,
    #[arg(long = "checks-grace-period")]
    checks_grace_period: Vec<String>,
    #[arg(long = "proxy-ports")]
    proxy_ports: Vec<String>,
    #[arg(long = "privileged")]
    privileged: bool,
    #[arg(long = "skip-cache")]
    skip_cache: bool,
    #[arg(long = "scale")]
    scale: Vec<String>,
    #[arg(long = "min-scale")]
    min_scale: Option<i64>,
    #[arg(long = "max-scale")]
    max_scale: Option<i64>,
    #[arg(long = "autoscaling-target")]
    autoscaling_target: Option<i64>,
    #[arg(long = "light-sleep-delay", default_value = "0s")]
    light_sleep_delay: HumanDuration,
    #[arg(long = "deep-sleep-delay", default_value = "0s")]
    deep_sleep_delay: HumanDuration,
    #[arg(long = "delete-after-delay", default_value = "0s")]
    delete_after_delay: HumanDuration,
    #[arg(long = "delete-after-inactivity-delay", default_value = "0s")]
    delete_after_inactivity_delay: HumanDuration,
}

#[derive(Debug, Args)]
struct LogsArgs {
    name: String,
    #[arg(short, long)]
    app: Option<String>,
    #[arg(long)]
    instance: Option<String>,
    #[arg(short = 't', long = "type")]
    log_type: Option<String>,
    #[arg(long)]
    since: Option<String>,
    #[arg(long)]
    tail: bool,
    #[arg(long = "start-time")]
    start_time: Option<String>,
    #[arg(long = "end-time")]
    end_time: Option<String>,
    #[arg(long = "regex-search")]
    regex_search: Option<String>,
    #[arg(long = "text-search")]
    text_search: Option<String>,
    #[arg(long, default_value = "asc")]
    order: String,
}

#[derive(Debug, Args)]
struct ScaleArgs {
    name: String,
    #[arg(short, long)]
    app: Option<String>,
    #[arg(long)]
    scale: Vec<String>,
    #[arg(long, default_value_t = 1)]
    instances: i64,
    #[arg(long)]
    regions: Vec<String>,
    #[command(subcommand)]
    command: Option<ScaleCommand>,
}

#[derive(Subcommand, Debug)]
enum ScaleCommand {
    #[command(
        about = "Update manual scaling configuration for service (patches existing configuration)"
    )]
    Update {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
        #[arg(long)]
        scale: Vec<String>,
        #[arg(long, default_value_t = 1)]
        instances: i64,
        #[arg(long)]
        regions: Vec<String>,
    },
    #[command(about = "Get manual scaling configuration for service")]
    Get {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
    },
    #[command(about = "Delete manual scaling configuration for service")]
    Delete {
        name: String,
        #[arg(short, long)]
        app: Option<String>,
    },
}

#[derive(Subcommand, Debug)]
enum DeploymentsCommand {
    #[command(about = "List deployments")]
    List {
        #[arg(long)]
        app: Option<String>,
        #[arg(long)]
        service: Option<String>,
    },
    #[command(about = "Get deployment")]
    Get { name: String },
    #[command(about = "Describe deployment")]
    Describe { name: String },
    #[command(about = "Cancel deployment")]
    Cancel { name: String },
    #[command(about = "Get deployment logs")]
    Logs(DeploymentLogsArgs),
}

#[derive(Debug, Args)]
struct DeploymentLogsArgs {
    name: String,
    #[arg(long)]
    instance: Option<String>,
    #[arg(short = 't', long = "type")]
    log_type: Option<String>,
    #[arg(long)]
    tail: bool,
    #[arg(long = "start-time")]
    start_time: Option<String>,
    #[arg(long = "end-time")]
    end_time: Option<String>,
    #[arg(long = "regex-search")]
    regex_search: Option<String>,
    #[arg(long = "text-search")]
    text_search: Option<String>,
    #[arg(long, default_value = "asc")]
    order: String,
}

#[derive(Subcommand, Debug)]
enum RegionalDeploymentsCommand {
    #[command(about = "List regional deployments")]
    List {
        #[arg(long)]
        deployment: Option<String>,
    },
    #[command(about = "Get regional deployment")]
    Get { name: String },
}

#[derive(Subcommand, Debug)]
enum InstancesCommand {
    #[command(about = "List instances")]
    List {
        #[arg(long)]
        app: Option<String>,
        #[arg(long)]
        service: Option<String>,
        #[arg(long)]
        deployment: Option<String>,
        #[arg(long)]
        regional_deployment: Option<String>,
    },
    #[command(about = "Get instance")]
    Get { name: String },
    #[command(about = "Describe instance")]
    Describe { name: String },
    #[command(about = "Get instance logs")]
    Logs(InstanceLogsArgs),
    #[command(
        about = "Run a command in the context of an instance",
        trailing_var_arg = true
    )]
    Exec {
        name: String,
        cmd: String,
        args: Vec<String>,
    },
    #[command(about = "Copy files and directories to and from instances.")]
    Cp { source: String, destination: String },
}

#[derive(Debug, Args)]
struct InstanceLogsArgs {
    name: String,
    #[arg(short = 't', long = "type")]
    log_type: Option<String>,
    #[arg(long)]
    tail: bool,
    #[arg(long = "start-time")]
    start_time: Option<String>,
    #[arg(long = "end-time")]
    end_time: Option<String>,
    #[arg(long = "regex-search")]
    regex_search: Option<String>,
    #[arg(long = "text-search")]
    text_search: Option<String>,
    #[arg(long, default_value = "asc")]
    order: String,
}

#[derive(Subcommand, Debug)]
enum DatabasesCommand {
    #[command(about = "List databases")]
    List,
    #[command(about = "Get database")]
    Get {
        name: String,
        #[arg(long)]
        app: Option<String>,
    },
    #[command(about = "Create database")]
    Create(DatabaseCreateArgs),
    #[command(about = "Update database")]
    Update(DatabaseUpdateArgs),
    #[command(about = "Delete database")]
    Delete { name: String },
}

#[derive(Debug, Args)]
struct DatabaseCreateArgs {
    name: String,
    #[arg(long)]
    app: Option<String>,
    #[arg(long = "pg-version")]
    pg_version: Option<i64>,
    #[arg(long)]
    region: Option<String>,
    #[arg(long = "instance-type")]
    instance_type: Option<String>,
    #[arg(long = "db-owner")]
    db_owner: Option<String>,
    #[arg(long = "db-name")]
    db_name: Option<String>,
}

#[derive(Debug, Args)]
struct DatabaseUpdateArgs {
    name: String,
    #[arg(long = "name")]
    new_name: Option<String>,
    #[arg(long = "pg-version")]
    pg_version: Option<i64>,
    #[arg(long)]
    region: Option<String>,
    #[arg(long = "instance-type")]
    instance_type: Option<String>,
    #[arg(long = "db-owner")]
    db_owner: Option<String>,
    #[arg(long = "db-name")]
    db_name: Option<String>,
}

#[derive(Subcommand, Debug)]
enum MetricsCommand {
    #[command(about = "Get metrics for a service or instance")]
    Get {
        #[arg(long)]
        service: Option<String>,
        #[arg(long)]
        instance: Option<String>,
        #[arg(long)]
        start: Option<String>,
        #[arg(long)]
        end: Option<String>,
    },
}

#[derive(Subcommand, Debug)]
enum VolumesCommand {
    #[command(about = "Create a new volume")]
    Create {
        name: String,
        #[arg(long, default_value = "fra")]
        region: String,
        #[arg(long, default_value_t = -1)]
        size: i64,
        #[arg(long)]
        read_only: bool,
        #[arg(long)]
        snapshot: Option<String>,
    },
    #[command(about = "Get a volume")]
    Get { name: String },
    #[command(about = "List volumes")]
    List,
    #[command(about = "Update a volume")]
    Update {
        name: String,
        #[arg(long = "name")]
        new_name: Option<String>,
        #[arg(long, default_value_t = -1)]
        size: i64,
    },
    #[command(about = "Delete a volume")]
    Delete { name: String },
}

#[derive(Subcommand, Debug)]
enum SnapshotsCommand {
    #[command(about = "Create a new snapshot")]
    Create { name: String, parent_volume: String },
    #[command(about = "Get a snapshot")]
    Get { name: String },
    #[command(about = "List snapshots")]
    List,
    #[command(about = "Update a snapshot")]
    Update {
        name: String,
        #[arg(long = "name")]
        new_name: Option<String>,
    },
    #[command(about = "Delete a snapshot")]
    Delete { name: String },
}

#[derive(Debug, Args)]
struct ComposeArgs {
    koyeb_compose_file_path: PathBuf,
    #[arg(short, long)]
    verbose: bool,
    #[command(subcommand)]
    command: Option<ComposeCommand>,
}

#[derive(Subcommand, Debug)]
enum ComposeCommand {
    #[command(about = "d")]
    Delete { koyeb_compose_file_path: PathBuf },
    #[command(about = "l")]
    Logs { koyeb_compose_file_path: PathBuf },
}

#[derive(Subcommand, Debug)]
enum SandboxCommand {
    #[command(about = "List sandboxes")]
    List {
        #[arg(short, long)]
        app: Option<String>,
        #[arg(short, long)]
        name: Option<String>,
    },
    #[command(about = "Create a new sandbox")]
    Create(SandboxCreateArgs),
    #[command(about = "Execute a command in the sandbox", trailing_var_arg = true)]
    Run {
        name: String,
        command: String,
        args: Vec<String>,
        #[arg(long)]
        cwd: Option<String>,
        #[arg(long)]
        env: Vec<String>,
        #[arg(long, default_value_t = 60)]
        timeout: i64,
        #[arg(long)]
        stream: bool,
    },
    #[command(
        visible_alias = "launch",
        about = "Start a background process in the sandbox",
        trailing_var_arg = true
    )]
    Start {
        name: String,
        command: String,
        args: Vec<String>,
        #[arg(long)]
        cwd: Option<String>,
        #[arg(long)]
        env: Vec<String>,
    },
    #[command(
        visible_alias = "list-processes",
        about = "List background processes in the sandbox"
    )]
    Ps { name: String },
    #[command(about = "Kill a background process in the sandbox")]
    Kill { name: String, process_id: String },
    #[command(about = "Stream logs from a background process")]
    Logs {
        name: String,
        process_id: String,
        #[arg(short, long)]
        follow: bool,
    },
    #[command(about = "Expose a port from the sandbox via TCP proxy")]
    ExposePort { name: String, port: u16 },
    #[command(about = "Unexpose the currently exposed port")]
    UnexposePort { name: String },
    #[command(about = "Check sandbox health status")]
    Health { name: String },
    #[command(visible_alias = "filesystem", about = "Filesystem operations")]
    Fs {
        #[command(subcommand)]
        command: SandboxFsCommand,
    },
}

#[derive(Debug, Args)]
struct SandboxCreateArgs {
    name: String,
    #[arg(long = "wait-timeout", default_value = "5m")]
    wait_timeout: HumanDuration,
    #[arg(long = "docker")]
    docker: Option<String>,
    #[arg(long = "docker-entrypoint")]
    docker_entrypoint: Vec<String>,
    #[arg(long = "docker-args")]
    docker_args: Vec<String>,
    #[arg(long = "regions")]
    regions: Vec<String>,
    #[arg(long = "env")]
    env: Vec<String>,
    #[arg(long = "config-file")]
    config_file: Vec<String>,
    #[arg(long = "delete-after-delay", default_value = "0s")]
    delete_after_delay: HumanDuration,
    #[arg(long = "delete-after-inactivity-delay", default_value = "0s")]
    delete_after_inactivity_delay: HumanDuration,
    #[arg(long = "light-sleep-delay", default_value = "0s")]
    light_sleep_delay: HumanDuration,
    #[arg(long = "deep-sleep-delay", default_value = "0s")]
    deep_sleep_delay: HumanDuration,
}

#[derive(Subcommand, Debug)]
enum SandboxFsCommand {
    #[command(about = "Read a file from the sandbox")]
    Read { name: String, path: String },
    #[command(about = "Write content to a file in the sandbox")]
    Write {
        name: String,
        path: String,
        content: Option<String>,
        #[arg(short, long)]
        file: Option<PathBuf>,
    },
    #[command(about = "List directory contents in the sandbox")]
    Ls {
        name: String,
        path: Option<String>,
        #[arg(short, long)]
        long: bool,
    },
    #[command(about = "Create a directory in the sandbox")]
    Mkdir { name: String, path: String },
    #[command(about = "Remove a file or directory from the sandbox")]
    Rm {
        name: String,
        path: String,
        #[arg(short, long)]
        recursive: bool,
    },
    #[command(about = "Upload a local file or directory to the sandbox (max 1G per file)")]
    Upload {
        name: String,
        local_path: PathBuf,
        remote_path: String,
        #[arg(short, long)]
        recursive: bool,
        #[arg(short, long)]
        force: bool,
    },
    #[command(about = "Download a file from the sandbox")]
    Download {
        name: String,
        remote_path: String,
        local_path: PathBuf,
    },
}

#[derive(Clone, Debug)]
struct HumanDuration(Duration);

impl Default for HumanDuration {
    fn default() -> Self {
        Self(Duration::ZERO)
    }
}
impl std::str::FromStr for HumanDuration {
    type Err = String;
    fn from_str(s: &str) -> std::result::Result<Self, Self::Err> {
        parse_duration(s).map(HumanDuration)
    }
}

fn parse_duration(s: &str) -> std::result::Result<Duration, String> {
    if s.is_empty() {
        return Err("duration cannot be empty".into());
    }
    let mut digits = String::new();
    let mut total = 0u64;
    for ch in s.chars() {
        if ch.is_ascii_digit() {
            digits.push(ch);
            continue;
        }
        let n: u64 = digits
            .parse()
            .map_err(|_| format!("invalid duration {s}"))?;
        digits.clear();
        total += match ch {
            's' => n,
            'm' => n * 60,
            'h' => n * 3600,
            'd' => n * 86400,
            _ => return Err(format!("invalid duration unit {ch}")),
        };
    }
    if !digits.is_empty() {
        total += digits
            .parse::<u64>()
            .map_err(|_| format!("invalid duration {s}"))?;
    }
    Ok(Duration::from_secs(total))
}

#[derive(Debug, Clone, Default)]
struct ConfigFile {
    token: Option<String>,
    url: Option<String>,
    organization: Option<String>,
}

#[derive(Debug, Clone)]
struct Config {
    url: String,
    token: Option<String>,
    organization: Option<String>,
    output: OutputFormat,
    debug: bool,
    debug_full: bool,
    full: bool,
    _force_ascii: bool,
    config_path: PathBuf,
}

impl Config {
    fn from_cli(cli: &Cli) -> Result<Self> {
        let path = config_path(cli.config.as_ref())?;
        let loaded = if path.exists() {
            load_config(&path)?
        } else {
            ConfigFile::default()
        };
        Ok(Self {
            url: if cli.url != DEFAULT_API_URL {
                cli.url.clone()
            } else {
                loaded.url.unwrap_or_else(|| DEFAULT_API_URL.to_string())
            },
            token: cli.token.clone().or(loaded.token),
            organization: cli.organization.clone().or(loaded.organization),
            output: cli.output,
            debug: cli.debug,
            debug_full: cli.debug_full,
            full: cli.full,
            _force_ascii: cli._force_ascii,
            config_path: path,
        })
    }
}

fn config_path(flag: Option<&PathBuf>) -> Result<PathBuf> {
    if let Some(path) = flag {
        return Ok(path.clone());
    }
    if let Ok(path) = std::env::var("KOYEB_CONFIG") {
        return Ok(PathBuf::from(path));
    }
    let home = dirs::home_dir()
        .ok_or_else(|| anyhow!("unable to find your home directory; set HOME or pass --config"))?;
    Ok(home.join(".koyeb.yaml"))
}

fn load_config(path: &Path) -> Result<ConfigFile> {
    let content = fs::read_to_string(path)
        .with_context(|| format!("unable to read config file {}", path.display()))?;
    let mut cfg = ConfigFile::default();
    for line in content.lines() {
        let line = line.trim();
        if line.is_empty() || line.starts_with('#') || line.starts_with('[') {
            continue;
        }
        let Some((key, value)) = line.split_once([':', '=']) else {
            continue;
        };
        let value = value
            .trim()
            .trim_matches('"')
            .trim_matches('\'')
            .to_string();
        match key.trim() {
            "token" => cfg.token = Some(value),
            "url" => cfg.url = Some(value),
            "organization" => cfg.organization = Some(value),
            _ => {}
        }
    }
    Ok(cfg)
}

fn save_config(path: &Path, config: &ConfigFile) -> Result<()> {
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
    }
    let mut data = String::new();
    if let Some(token) = &config.token {
        data.push_str(&format!("token: {}\n", token));
    }
    if let Some(url) = &config.url {
        data.push_str(&format!("url: {}\n", url));
    }
    if let Some(organization) = &config.organization {
        data.push_str(&format!("organization: {}\n", organization));
    }
    fs::write(path, data).with_context(|| format!("unable to write config file {}", path.display()))
}

pub fn run_blocking() -> Result<()> {
    let rt = tokio::runtime::Runtime::new()?;
    rt.block_on(run())
}

pub async fn run() -> Result<()> {
    let cli = Cli::parse();
    execute(cli).await
}

async fn execute(cli: Cli) -> Result<()> {
    match &cli.command {
        Some(Commands::Completion { shell }) => return print_completion(shell.clone()),
        Some(Commands::Version) => {
            println!("{VERSION}");
            return Ok(());
        }
        _ => {}
    }

    let cfg = Config::from_cli(&cli)?;
    match cli.command.unwrap_or_else(|| {
        let _ = Cli::command().print_help();
        std::process::exit(0)
    }) {
        Commands::Login => login(cfg).await,
        Commands::Apps { command } => handle_apps(&cfg, command).await,
        Commands::Archives { command } => handle_archives(&cfg, command).await,
        Commands::Deploy(args) => handle_deploy(&cfg, args).await,
        Commands::Domains { command } => handle_domains(&cfg, command).await,
        Commands::Organizations { command } => handle_organizations(&cfg, command).await,
        Commands::Secrets { command } => handle_secrets(&cfg, command).await,
        Commands::Services { command } => handle_services(&cfg, command).await,
        Commands::Deployments { command } => handle_deployments(&cfg, command).await,
        Commands::RegionalDeployments { command } => {
            handle_regional_deployments(&cfg, command).await
        }
        Commands::Instances { command } => handle_instances(&cfg, command).await,
        Commands::Databases { command } => handle_databases(&cfg, command).await,
        Commands::Metrics { command } => handle_metrics(&cfg, command).await,
        Commands::Volumes { command } => handle_volumes(&cfg, command).await,
        Commands::Snapshots { command } => handle_snapshots(&cfg, command).await,
        Commands::Compose(args) => handle_compose(&cfg, args).await,
        Commands::Sandbox { command } => handle_sandbox(&cfg, command).await,
        Commands::Whoami => handle_whoami(&cfg).await,
        Commands::Version | Commands::Completion { .. } => unreachable!(),
    }
}

fn print_completion(shell: CompletionShell) -> Result<()> {
    let mut cmd = Cli::command();
    let name = cmd.get_name().to_string();
    match shell {
        CompletionShell::Bash => clap_complete::generate(
            clap_complete::Shell::Bash,
            &mut cmd,
            name,
            &mut io::stdout(),
        ),
        CompletionShell::Zsh => {
            clap_complete::generate(clap_complete::Shell::Zsh, &mut cmd, name, &mut io::stdout())
        }
        CompletionShell::Fish => clap_complete::generate(
            clap_complete::Shell::Fish,
            &mut cmd,
            name,
            &mut io::stdout(),
        ),
        CompletionShell::Powershell => clap_complete::generate(
            clap_complete::Shell::PowerShell,
            &mut cmd,
            name,
            &mut io::stdout(),
        ),
    }
    Ok(())
}

async fn login(cfg: Config) -> Result<()> {
    print!(
        "Enter your personal access token. You can create a new token here (https://app.koyeb.com/user/settings/api): "
    );
    io::stdout().flush()?;
    let token = rpassword::read_password().or_else(|_| {
        let mut s = String::new();
        io::stdin().read_line(&mut s)?;
        Ok::<_, io::Error>(s.trim().to_string())
    })?;
    if token.trim().is_empty() {
        bail!("token cannot be empty");
    }
    let config = ConfigFile {
        token: Some(token.trim().to_string()),
        url: Some(cfg.url),
        organization: cfg.organization,
    };
    save_config(&cfg.config_path, &config)?;
    println!("Configuration written to {}", cfg.config_path.display());
    Ok(())
}

#[derive(Clone)]
struct ApiClient {
    cfg: Config,
    http: reqwest::Client,
}

impl ApiClient {
    fn new(cfg: &Config) -> Result<Self> {
        Ok(Self {
            cfg: cfg.clone(),
            http: reqwest::Client::builder()
                .user_agent(format!("koyeb-cli/{VERSION}"))
                .build()?,
        })
    }

    async fn request(
        &self,
        method: reqwest::Method,
        path: &str,
        body: Option<Value>,
        query: &[(&str, String)],
    ) -> Result<Value> {
        let base = self.cfg.url.trim_end_matches('/');
        let url = if path.starts_with("http") {
            path.to_string()
        } else {
            format!("{base}{path}")
        };
        let mut req = self.http.request(method.clone(), &url);
        if let Some(token) = &self.cfg.token {
            req = req.bearer_auth(token);
        }
        if let Some(org) = &self.cfg.organization {
            req = req.header("X-Koyeb-Organization", org);
        }
        if !query.is_empty() {
            req = req.query(query);
        }
        if let Some(body) = body {
            req = req.json(&body);
        }
        if self.cfg.debug {
            eprintln!("{} {}", method, url);
            if self.cfg.debug_full {
                eprintln!("debug-full enabled: request headers include authorization");
            }
        }
        let resp = req
            .send()
            .await
            .with_context(|| format!("request to {url} failed"))?;
        let status = resp.status();
        let text = resp.text().await.unwrap_or_default();
        if !status.is_success() {
            let msg = serde_json::from_str::<Value>(&text)
                .ok()
                .and_then(|v| v.get("message").and_then(Value::as_str).map(str::to_string))
                .unwrap_or(text);
            bail!("API request failed ({status}): {msg}");
        }
        if text.trim().is_empty() {
            Ok(Value::Null)
        } else {
            Ok(serde_json::from_str(&text).unwrap_or_else(|_| Value::String(text)))
        }
    }

    async fn get(&self, path: &str, query: &[(&str, String)]) -> Result<Value> {
        self.request(reqwest::Method::GET, path, None, query).await
    }
    async fn post(&self, path: &str, body: Value) -> Result<Value> {
        self.request(reqwest::Method::POST, path, Some(body), &[])
            .await
    }
    async fn put(&self, path: &str, body: Value) -> Result<Value> {
        self.request(reqwest::Method::PUT, path, Some(body), &[])
            .await
    }
    async fn patch(&self, path: &str, body: Value) -> Result<Value> {
        self.request(reqwest::Method::PATCH, path, Some(body), &[])
            .await
    }
    async fn delete(&self, path: &str) -> Result<Value> {
        self.request(reqwest::Method::DELETE, path, None, &[]).await
    }
}

fn auth_client(cfg: &Config) -> Result<ApiClient> {
    if cfg.token.as_deref().unwrap_or_default().is_empty() {
        bail!("missing API token; run `koyeb login`, pass --token, or set KOYEB_TOKEN");
    }
    ApiClient::new(cfg)
}

async fn handle_apps(cfg: &Config, command: AppsCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        AppsCommand::Create {
            name,
            delete_when_empty,
        } => render(
            cfg,
            api.post(
                "/v1/apps",
                json!({"app": {"name": name, "delete_when_empty": delete_when_empty}}),
            )
            .await?,
        ),
        AppsCommand::Get { name } | AppsCommand::Describe { name } => render(
            cfg,
            api.get(
                &format!("/v1/apps/{}", encode(&resolve(&api, "apps", &name).await?)),
                &[],
            )
            .await?,
        ),
        AppsCommand::List => render(cfg, api.get("/v1/apps", &[limit()]).await?),
        AppsCommand::Update {
            name,
            name_new,
            domain,
            delete_when_empty,
        } => {
            let id = resolve(&api, "apps", &name).await?;
            let mut app = json!({});
            if let Some(v) = name_new {
                app["name"] = json!(v);
            }
            if let Some(v) = domain {
                app["domains"] = json!([{ "name": v }]);
            }
            if delete_when_empty {
                app["delete_when_empty"] = json!(true);
            }
            render(
                cfg,
                api.patch(&format!("/v1/apps/{}", encode(&id)), json!({"app": app}))
                    .await?,
            )
        }
        AppsCommand::Delete { name } => {
            let id = resolve(&api, "apps", &name).await?;
            render(cfg, api.delete(&format!("/v1/apps/{}", encode(&id))).await?)
        }
        AppsCommand::Pause { name } => {
            let id = resolve(&api, "apps", &name).await?;
            render(
                cfg,
                api.post(&format!("/v1/apps/{}/pause", encode(&id)), json!({}))
                    .await?,
            )
        }
        AppsCommand::Resume { name } => {
            let id = resolve(&api, "apps", &name).await?;
            render(
                cfg,
                api.post(&format!("/v1/apps/{}/resume", encode(&id)), json!({}))
                    .await?,
            )
        }
        AppsCommand::Init(args) => {
            let app = api
                .post("/v1/apps", json!({"app": {"name": args.name}}))
                .await?;
            let app_id = app
                .pointer("/app/id")
                .and_then(Value::as_str)
                .unwrap_or_default()
                .to_string();
            let body = service_body(&args.name, args.service, Some(app_id), false)?;
            let svc = api.post("/v1/services", body).await?;
            if args.wait {
                wait_for_service(&api, &svc, args.wait_timeout.0).await?;
            }
            render(cfg, svc)
        }
    }
}

async fn handle_archives(cfg: &Config, command: ArchivesCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        ArchivesCommand::Create { name, ignore_dir } => {
            let archive = create_tarball(Path::new("."), &ignore_dir)?;
            let encoded =
                base64::Engine::encode(&base64::engine::general_purpose::STANDARD, archive);
            render(
                cfg,
                api.post(
                    "/v1/archives",
                    json!({"archive": {"name": name, "data": encoded}}),
                )
                .await?,
            )
        }
    }
}

async fn handle_deploy(cfg: &Config, args: DeployArgs) -> Result<()> {
    let api = auth_client(cfg)?;
    let (app_name, service_name) = split_app_service(args.app.as_deref(), &args.target)?;
    let app_id = ensure_app(&api, &app_name).await?;
    let archive = create_tarball(
        &args.path,
        &[".git".into(), "node_modules".into(), "vendor".into()],
    )?;
    let archive_resp = api.post("/v1/archives", json!({"archive": {"name": format!("{}-archive", service_name), "data": base64::Engine::encode(&base64::engine::general_purpose::STANDARD, archive)}})).await?;
    let archive_id = archive_resp
        .pointer("/archive/id")
        .and_then(Value::as_str)
        .unwrap_or_default()
        .to_string();
    let mut def = args.service;
    def.archive = Some(archive_id);
    let body = service_body(&service_name, def, Some(app_id), false)?;
    let res = match api.post("/v1/services", body).await {
        Ok(value) => value,
        Err(_) => {
            let id = resolve_scoped(&api, "services", &service_name, Some((&app_name, "app_id")))
                .await?;
            api.patch(
                &format!("/v1/services/{}", encode(&id)),
                service_body(&service_name, ServiceDefinitionArgs::default(), None, false)?,
            )
            .await?
        }
    };
    if args.wait {
        wait_for_service(&api, &res, args.wait_timeout.0).await?;
    }
    render(cfg, res)
}

async fn handle_domains(cfg: &Config, command: DomainsCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        DomainsCommand::Get { name } | DomainsCommand::Describe { name } => {
            let id = resolve(&api, "domains", &name).await?;
            render(
                cfg,
                api.get(&format!("/v1/domains/{}", encode(&id)), &[])
                    .await?,
            )
        }
        DomainsCommand::Create { name, attach_to } => {
            let mut domain = json!({"name": name});
            if let Some(app) = attach_to {
                domain["app_id"] = json!(resolve(&api, "apps", &app).await?);
            }
            render(
                cfg,
                api.post("/v1/domains", json!({"domain": domain})).await?,
            )
        }
        DomainsCommand::List => render(cfg, api.get("/v1/domains", &[limit()]).await?),
        DomainsCommand::Delete { name } => {
            let id = resolve(&api, "domains", &name).await?;
            render(
                cfg,
                api.delete(&format!("/v1/domains/{}", encode(&id))).await?,
            )
        }
        DomainsCommand::Refresh { name } => {
            let id = resolve(&api, "domains", &name).await?;
            render(
                cfg,
                api.post(&format!("/v1/domains/{}/refresh", encode(&id)), json!({}))
                    .await?,
            )
        }
        DomainsCommand::Attach { name, app } => {
            let id = resolve(&api, "domains", &name).await?;
            let app_id = resolve(&api, "apps", &app).await?;
            render(
                cfg,
                api.patch(
                    &format!("/v1/domains/{}", encode(&id)),
                    json!({"domain": {"app_id": app_id}}),
                )
                .await?,
            )
        }
        DomainsCommand::Detach { name } => {
            let id = resolve(&api, "domains", &name).await?;
            render(
                cfg,
                api.patch(
                    &format!("/v1/domains/{}", encode(&id)),
                    json!({"domain": {"app_id": null}}),
                )
                .await?,
            )
        }
    }
}

async fn handle_organizations(cfg: &Config, command: OrganizationsCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        OrganizationsCommand::List => {
            let value = match api.get("/v1/organizations", &[limit()]).await {
                Ok(value) => value,
                Err(_) => api.get("/v1/profile/organizations", &[limit()]).await?,
            };
            render(cfg, value)
        }
        OrganizationsCommand::Switch { organization } => {
            let mut file = if cfg.config_path.exists() {
                load_config(&cfg.config_path)?
            } else {
                ConfigFile::default()
            };
            file.organization = Some(organization.clone());
            if file.token.is_none() {
                file.token = cfg.token.clone();
            }
            if file.url.is_none() {
                file.url = Some(cfg.url.clone());
            }
            save_config(&cfg.config_path, &file)?;
            println!("Switched to organization {organization}");
            Ok(())
        }
    }
}

async fn handle_secrets(cfg: &Config, command: SecretsCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        SecretsCommand::Create {
            name,
            kind,
            value,
            filename,
            registry_server,
            registry_username,
            registry_password,
        } => {
            let mut secret = json!({"name": name});
            match kind {
                SecretKind::Simple => secret["value"] = json!(read_secret_value(value, filename)?),
                SecretKind::Registry => {
                    secret["type"] = json!("REGISTRY");
                    secret["docker_hub_registry"] = json!({"server": registry_server, "username": registry_username, "password": registry_password});
                }
            }
            render(
                cfg,
                api.post("/v1/secrets", json!({"secret": secret})).await?,
            )
        }
        SecretsCommand::Get { name } | SecretsCommand::Describe { name } => {
            let id = resolve(&api, "secrets", &name).await?;
            render(
                cfg,
                api.get(&format!("/v1/secrets/{}", encode(&id)), &[])
                    .await?,
            )
        }
        SecretsCommand::List => render(cfg, api.get("/v1/secrets", &[limit()]).await?),
        SecretsCommand::Update {
            name,
            value,
            filename,
        } => {
            let id = resolve(&api, "secrets", &name).await?;
            render(
                cfg,
                api.patch(
                    &format!("/v1/secrets/{}", encode(&id)),
                    json!({"secret": {"value": read_secret_value(value, filename)?}}),
                )
                .await?,
            )
        }
        SecretsCommand::Delete { name } => {
            let id = resolve(&api, "secrets", &name).await?;
            render(
                cfg,
                api.delete(&format!("/v1/secrets/{}", encode(&id))).await?,
            )
        }
        SecretsCommand::Reveal { name } => {
            let id = resolve(&api, "secrets", &name).await?;
            render(
                cfg,
                api.post(&format!("/v1/secrets/{}/reveal", encode(&id)), json!({}))
                    .await?,
            )
        }
    }
}

async fn handle_services(cfg: &Config, command: ServicesCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        ServicesCommand::Create(args) => {
            let app_id = match args.app.as_deref() {
                Some(app) => Some(ensure_app(&api, app).await?),
                None => None,
            };
            let body = service_body(&args.name, args.definition, app_id, false)?;
            let res = api.post("/v1/services", body).await?;
            if args.wait {
                wait_for_service(&api, &res, args.wait_timeout.0).await?;
            }
            render(cfg, res)
        }
        ServicesCommand::Get { name, app } | ServicesCommand::Describe { name, app } => {
            let id = resolve_scoped(
                &api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.get(&format!("/v1/services/{}", encode(&id)), &[])
                    .await?,
            )
        }
        ServicesCommand::UnappliedChanges { service_name, app } => {
            let id = resolve_scoped(
                &api,
                "services",
                &service_name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.get(
                    &format!("/v1/services/{}/unapplied_changes", encode(&id)),
                    &[],
                )
                .await?,
            )
        }
        ServicesCommand::Logs(args) => service_logs(cfg, &api, args).await,
        ServicesCommand::List { app, name } => {
            let mut q = vec![limit()];
            if let Some(name) = name {
                q.push(("name", name));
            }
            if let Some(app) = app {
                q.push(("app_id", resolve(&api, "apps", &app).await?));
            }
            render(cfg, api.get("/v1/services", &q).await?)
        }
        ServicesCommand::Exec {
            name,
            cmd,
            args,
            app,
        } => exec_remote(&api, cfg, "services", &name, app.as_deref(), cmd, args).await,
        ServicesCommand::Update(args) => {
            let id = resolve_scoped(
                &api,
                "services",
                &args.name,
                args.app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            let mut body = service_body(
                args.new_name.as_deref().unwrap_or(&args.name),
                args.definition,
                None,
                args.override_,
            )?;
            body["skip_build"] = json!(args.skip_build);
            body["save_only"] = json!(args.save_only);
            let res = api
                .patch(&format!("/v1/services/{}", encode(&id)), body)
                .await?;
            if args.wait {
                wait_for_service(&api, &res, args.wait_timeout.0).await?;
            }
            render(cfg, res)
        }
        ServicesCommand::Redeploy {
            name,
            app,
            skip_build,
            wait,
            wait_timeout,
            use_cache,
        } => {
            let id = resolve_scoped(
                &api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            let res = api
                .post(
                    &format!("/v1/services/{}/redeploy", encode(&id)),
                    json!({"skip_build": skip_build, "use_cache": use_cache}),
                )
                .await?;
            if wait {
                wait_for_service(&api, &res, wait_timeout.0).await?;
            }
            render(cfg, res)
        }
        ServicesCommand::Delete { name, app } => {
            let id = resolve_scoped(
                &api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.delete(&format!("/v1/services/{}", encode(&id))).await?,
            )
        }
        ServicesCommand::Pause { name, app } => {
            let id = resolve_scoped(
                &api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.post(&format!("/v1/services/{}/pause", encode(&id)), json!({}))
                    .await?,
            )
        }
        ServicesCommand::Resume { name, app } => {
            let id = resolve_scoped(
                &api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.post(&format!("/v1/services/{}/resume", encode(&id)), json!({}))
                    .await?,
            )
        }
        ServicesCommand::Scale(args) => handle_scale(cfg, &api, args).await,
    }
}

async fn handle_scale(cfg: &Config, api: &ApiClient, args: ScaleArgs) -> Result<()> {
    match args.command {
        Some(ScaleCommand::Get { name, app }) => {
            let id = resolve_scoped(
                api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.get(&format!("/v1/services/{}/scaling", encode(&id)), &[])
                    .await?,
            )
        }
        Some(ScaleCommand::Delete { name, app }) => {
            let id = resolve_scoped(
                api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.delete(&format!("/v1/services/{}/scaling", encode(&id)))
                    .await?,
            )
        }
        Some(ScaleCommand::Update {
            name,
            app,
            scale,
            instances,
            regions,
        }) => {
            let id = resolve_scoped(
                api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.patch(
                    &format!("/v1/services/{}/scaling", encode(&id)),
                    scaling_body(scale, instances, regions),
                )
                .await?,
            )
        }
        None => {
            let id = resolve_scoped(
                api,
                "services",
                &args.name,
                args.app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.put(
                    &format!("/v1/services/{}/scaling", encode(&id)),
                    scaling_body(args.scale, args.instances, args.regions),
                )
                .await?,
            )
        }
    }
}

async fn handle_deployments(cfg: &Config, command: DeploymentsCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        DeploymentsCommand::List { app, service } => {
            let mut q = vec![limit()];
            if let Some(app) = app {
                q.push(("app_id", resolve(&api, "apps", &app).await?));
            }
            if let Some(service) = service {
                q.push(("service_id", resolve(&api, "services", &service).await?));
            }
            render(cfg, api.get("/v1/deployments", &q).await?)
        }
        DeploymentsCommand::Get { name } | DeploymentsCommand::Describe { name } => {
            let id = resolve(&api, "deployments", &name).await?;
            render(
                cfg,
                api.get(&format!("/v1/deployments/{}", encode(&id)), &[])
                    .await?,
            )
        }
        DeploymentsCommand::Cancel { name } => {
            let id = resolve(&api, "deployments", &name).await?;
            render(
                cfg,
                api.post(
                    &format!("/v1/deployments/{}/cancel", encode(&id)),
                    json!({}),
                )
                .await?,
            )
        }
        DeploymentsCommand::Logs(args) => {
            logs_query(
                cfg,
                &api,
                "deployment_id",
                &resolve(&api, "deployments", &args.name).await?,
                args.instance,
                args.log_type,
                args.tail,
                args.start_time,
                args.end_time,
                args.regex_search,
                args.text_search,
                args.order,
            )
            .await
        }
    }
}

async fn handle_regional_deployments(
    cfg: &Config,
    command: RegionalDeploymentsCommand,
) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        RegionalDeploymentsCommand::List { deployment } => {
            let mut q = vec![limit()];
            if let Some(d) = deployment {
                q.push(("deployment_id", resolve(&api, "deployments", &d).await?));
            }
            render(cfg, api.get("/v1/regional_deployments", &q).await?)
        }
        RegionalDeploymentsCommand::Get { name } => {
            let id = resolve(&api, "regional_deployments", &name).await?;
            render(
                cfg,
                api.get(&format!("/v1/regional_deployments/{}", encode(&id)), &[])
                    .await?,
            )
        }
    }
}

async fn handle_instances(cfg: &Config, command: InstancesCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        InstancesCommand::List {
            app,
            service,
            deployment,
            regional_deployment,
        } => {
            let mut q = vec![limit()];
            if let Some(v) = app {
                q.push(("app_id", resolve(&api, "apps", &v).await?));
            }
            if let Some(v) = service {
                q.push(("service_id", resolve(&api, "services", &v).await?));
            }
            if let Some(v) = deployment {
                q.push(("deployment_id", resolve(&api, "deployments", &v).await?));
            }
            if let Some(v) = regional_deployment {
                q.push((
                    "regional_deployment_id",
                    resolve(&api, "regional_deployments", &v).await?,
                ));
            }
            render(cfg, api.get("/v1/instances", &q).await?)
        }
        InstancesCommand::Get { name } | InstancesCommand::Describe { name } => {
            let id = resolve(&api, "instances", &name).await?;
            render(
                cfg,
                api.get(&format!("/v1/instances/{}", encode(&id)), &[])
                    .await?,
            )
        }
        InstancesCommand::Logs(args) => {
            logs_query(
                cfg,
                &api,
                "instance_id",
                &resolve(&api, "instances", &args.name).await?,
                None,
                args.log_type,
                args.tail,
                args.start_time,
                args.end_time,
                args.regex_search,
                args.text_search,
                args.order,
            )
            .await
        }
        InstancesCommand::Exec { name, cmd, args } => {
            exec_remote(&api, cfg, "instances", &name, None, cmd, args).await
        }
        InstancesCommand::Cp {
            source,
            destination,
        } => cp_remote(&api, cfg, source, destination).await,
    }
}

async fn handle_databases(cfg: &Config, command: DatabasesCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        DatabasesCommand::List => render(
            cfg,
            api.get("/v1/services", &[("type", "DATABASE".into()), limit()])
                .await?,
        ),
        DatabasesCommand::Get { name, app } => {
            let id = resolve_scoped(
                &api,
                "services",
                &name,
                app.as_deref().map(|a| (a, "app_id")),
            )
            .await?;
            render(
                cfg,
                api.get(&format!("/v1/services/{}", encode(&id)), &[])
                    .await?,
            )
        }
        DatabasesCommand::Create(args) => {
            let app_id = match args.app.as_deref() {
                Some(app) => Some(ensure_app(&api, app).await?),
                None => None,
            };
            render(
                cfg,
                api.post(
                    "/v1/services",
                    database_body(
                        args.name,
                        app_id,
                        args.pg_version,
                        args.region,
                        args.instance_type,
                        args.db_owner,
                        args.db_name,
                    ),
                )
                .await?,
            )
        }
        DatabasesCommand::Update(args) => {
            let id = resolve(&api, "services", &args.name).await?;
            render(
                cfg,
                api.patch(
                    &format!("/v1/services/{}", encode(&id)),
                    database_body(
                        args.new_name.unwrap_or(args.name),
                        None,
                        args.pg_version,
                        args.region,
                        args.instance_type,
                        args.db_owner,
                        args.db_name,
                    ),
                )
                .await?,
            )
        }
        DatabasesCommand::Delete { name } => {
            let id = resolve(&api, "services", &name).await?;
            render(
                cfg,
                api.delete(&format!("/v1/services/{}", encode(&id))).await?,
            )
        }
    }
}

async fn handle_metrics(cfg: &Config, command: MetricsCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        MetricsCommand::Get {
            service,
            instance,
            start,
            end,
        } => {
            if service.is_some() == instance.is_some() {
                bail!("you must specify exactly one of --service or --instance");
            }
            let mut q = vec![];
            if let Some(s) = service {
                q.push(("service_id", resolve(&api, "services", &s).await?));
            }
            if let Some(i) = instance {
                q.push(("instance_id", resolve(&api, "instances", &i).await?));
            }
            if let Some(s) = start {
                q.push(("start", s));
            }
            if let Some(e) = end {
                q.push(("end", e));
            }
            render(cfg, api.get("/v1/metrics", &q).await?)
        }
    }
}

async fn handle_volumes(cfg: &Config, command: VolumesCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        VolumesCommand::Create {
            name,
            region,
            size,
            read_only,
            snapshot,
        } => {
            let mut volume = json!({"name": name, "region": region});
            if size >= 0 {
                volume["size"] = json!(size);
            }
            if read_only {
                volume["read_only"] = json!(true);
            }
            if let Some(snapshot) = snapshot {
                volume["snapshot_id"] = json!(resolve(&api, "snapshots", &snapshot).await?);
            }
            render(
                cfg,
                api.post("/v1/persistent_volumes", json!({"volume": volume}))
                    .await?,
            )
        }
        VolumesCommand::Get { name } => {
            let id = resolve(&api, "persistent_volumes", &name).await?;
            render(
                cfg,
                api.get(&format!("/v1/persistent_volumes/{}", encode(&id)), &[])
                    .await?,
            )
        }
        VolumesCommand::List => render(cfg, api.get("/v1/persistent_volumes", &[limit()]).await?),
        VolumesCommand::Update {
            name,
            new_name,
            size,
        } => {
            let id = resolve(&api, "persistent_volumes", &name).await?;
            let mut v = json!({});
            if let Some(n) = new_name {
                v["name"] = json!(n);
            }
            if size >= 0 {
                v["size"] = json!(size);
            }
            render(
                cfg,
                api.patch(
                    &format!("/v1/persistent_volumes/{}", encode(&id)),
                    json!({"volume": v}),
                )
                .await?,
            )
        }
        VolumesCommand::Delete { name } => {
            let id = resolve(&api, "persistent_volumes", &name).await?;
            render(
                cfg,
                api.delete(&format!("/v1/persistent_volumes/{}", encode(&id)))
                    .await?,
            )
        }
    }
}

async fn handle_snapshots(cfg: &Config, command: SnapshotsCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        SnapshotsCommand::Create {
            name,
            parent_volume,
        } => {
            let parent_volume_id = resolve(&api, "persistent_volumes", &parent_volume).await?;
            render(
                cfg,
                api.post(
                    "/v1/snapshots",
                    json!({"snapshot": {"name": name, "parent_volume_id": parent_volume_id}}),
                )
                .await?,
            )
        }
        SnapshotsCommand::Get { name } => {
            let id = resolve(&api, "snapshots", &name).await?;
            render(
                cfg,
                api.get(&format!("/v1/snapshots/{}", encode(&id)), &[])
                    .await?,
            )
        }
        SnapshotsCommand::List => render(cfg, api.get("/v1/snapshots", &[limit()]).await?),
        SnapshotsCommand::Update { name, new_name } => {
            let id = resolve(&api, "snapshots", &name).await?;
            render(
                cfg,
                api.patch(
                    &format!("/v1/snapshots/{}", encode(&id)),
                    json!({"snapshot": {"name": new_name}}),
                )
                .await?,
            )
        }
        SnapshotsCommand::Delete { name } => {
            let id = resolve(&api, "snapshots", &name).await?;
            render(
                cfg,
                api.delete(&format!("/v1/snapshots/{}", encode(&id)))
                    .await?,
            )
        }
    }
}

async fn handle_compose(cfg: &Config, args: ComposeArgs) -> Result<()> {
    let api = auth_client(cfg)?;
    match args.command {
        Some(ComposeCommand::Delete {
            koyeb_compose_file_path,
        }) => {
            let compose = read_yaml(&koyeb_compose_file_path)?;
            render(cfg, api.post("/v1/compose/delete", compose).await?)
        }
        Some(ComposeCommand::Logs {
            koyeb_compose_file_path,
        }) => {
            let compose = read_yaml(&koyeb_compose_file_path)?;
            render(cfg, api.post("/v1/compose/logs", compose).await?)
        }
        None => {
            let compose = read_yaml(&args.koyeb_compose_file_path)?;
            let res = api.post("/v1/compose", compose).await?;
            if args.verbose {
                eprintln!("compose submitted");
            }
            render(cfg, res)
        }
    }
}

async fn handle_sandbox(cfg: &Config, command: SandboxCommand) -> Result<()> {
    let api = auth_client(cfg)?;
    match command {
        SandboxCommand::List { app, name } => { let mut q = vec![("type", "SANDBOX".into()), limit()]; if let Some(n) = name { q.push(("name", n)); } if let Some(app) = app { q.push(("app_id", resolve(&api, "apps", &app).await?)); } render(cfg, api.get("/v1/services", &q).await?) }
        SandboxCommand::Create(args) => {
            let body = sandbox_body(args)?;
            let value = match api.post("/v1/sandboxes", body.clone()).await {
                Ok(value) => value,
                Err(_) => api.post("/v1/services", body).await?,
            };
            render(cfg, value)
        }
        SandboxCommand::Run { name, command, args, cwd, env, timeout, stream } => sandbox_rpc(cfg, &api, &name, "run", json!({"command": command, "args": args, "cwd": cwd, "env": parse_key_values(env)?, "timeout": timeout, "stream": stream})).await,
        SandboxCommand::Start { name, command, args, cwd, env } => sandbox_rpc(cfg, &api, &name, "processes", json!({"command": command, "args": args, "cwd": cwd, "env": parse_key_values(env)?})).await,
        SandboxCommand::Ps { name } => sandbox_rpc(cfg, &api, &name, "processes", json!({})).await,
        SandboxCommand::Kill { name, process_id } => sandbox_rpc(cfg, &api, &name, &format!("processes/{process_id}/kill"), json!({})).await,
        SandboxCommand::Logs { name, process_id, follow } => sandbox_rpc(cfg, &api, &name, &format!("processes/{process_id}/logs"), json!({"follow": follow})).await,
        SandboxCommand::ExposePort { name, port } => sandbox_rpc(cfg, &api, &name, "ports/expose", json!({"port": port})).await,
        SandboxCommand::UnexposePort { name } => sandbox_rpc(cfg, &api, &name, "ports/unexpose", json!({})).await,
        SandboxCommand::Health { name } => sandbox_rpc(cfg, &api, &name, "health", json!({})).await,
        SandboxCommand::Fs { command } => handle_sandbox_fs(cfg, &api, command).await,
    }
}

async fn handle_sandbox_fs(cfg: &Config, api: &ApiClient, command: SandboxFsCommand) -> Result<()> {
    match command {
        SandboxFsCommand::Read { name, path } => {
            sandbox_rpc(cfg, api, &name, "fs/read", json!({"path": path})).await
        }
        SandboxFsCommand::Write {
            name,
            path,
            content,
            file,
        } => {
            let content = match (content, file) {
                (Some(c), None) => c,
                (_, Some(p)) => fs::read_to_string(p)?,
                (None, None) => {
                    let mut s = String::new();
                    io::stdin().read_to_string(&mut s)?;
                    s
                }
            };
            sandbox_rpc(
                cfg,
                api,
                &name,
                "fs/write",
                json!({"path": path, "content": content}),
            )
            .await
        }
        SandboxFsCommand::Ls { name, path, long } => {
            sandbox_rpc(
                cfg,
                api,
                &name,
                "fs/ls",
                json!({"path": path.unwrap_or_else(|| ".".into()), "long": long}),
            )
            .await
        }
        SandboxFsCommand::Mkdir { name, path } => {
            sandbox_rpc(cfg, api, &name, "fs/mkdir", json!({"path": path})).await
        }
        SandboxFsCommand::Rm {
            name,
            path,
            recursive,
        } => {
            sandbox_rpc(
                cfg,
                api,
                &name,
                "fs/rm",
                json!({"path": path, "recursive": recursive}),
            )
            .await
        }
        SandboxFsCommand::Upload {
            name,
            local_path,
            remote_path,
            recursive,
            force,
        } => {
            let data = if local_path.is_dir() {
                base64::Engine::encode(
                    &base64::engine::general_purpose::STANDARD,
                    create_tarball(&local_path, &[])?,
                )
            } else {
                base64::Engine::encode(
                    &base64::engine::general_purpose::STANDARD,
                    fs::read(&local_path)?,
                )
            };
            sandbox_rpc(cfg, api, &name, "fs/upload", json!({"remote_path": remote_path, "data": data, "recursive": recursive, "force": force})).await
        }
        SandboxFsCommand::Download {
            name,
            remote_path,
            local_path,
        } => {
            let res =
                sandbox_rpc_value(api, &name, "fs/download", json!({"path": remote_path})).await?;
            if let Some(data) = res.get("data").and_then(Value::as_str) {
                fs::write(
                    local_path,
                    base64::Engine::decode(&base64::engine::general_purpose::STANDARD, data)?,
                )?;
                Ok(())
            } else {
                render(cfg, res)
            }
        }
    }
}

async fn handle_whoami(cfg: &Config) -> Result<()> {
    let api = auth_client(cfg)?;
    let value = match api.get("/v1/profile", &[]).await {
        Ok(value) => value,
        Err(_) => api.get("/v1/account", &[]).await?,
    };
    render(cfg, value)
}

fn render(cfg: &Config, value: Value) -> Result<()> {
    match cfg.output {
        OutputFormat::Json => println!("{}", serde_json::to_string_pretty(&value)?),
        OutputFormat::Yaml => println!("{}", serde_yaml::to_string(&value)?),
        OutputFormat::Table => print_table(&value, cfg.full)?,
    }
    Ok(())
}

fn print_table(value: &Value, full: bool) -> Result<()> {
    match value {
        Value::Null => println!("OK"),
        Value::Array(items) => print_rows(items, full),
        Value::Object(map) => {
            if let Some(arr) = map.values().find_map(Value::as_array) {
                print_rows(arr, full);
            } else if let Some(obj) = map.values().find_map(Value::as_object) {
                print_object(obj, full);
            } else {
                print_object(map, full);
            }
        }
        _ => println!("{value}"),
    }
    Ok(())
}

fn print_rows(items: &[Value], full: bool) {
    if items.is_empty() {
        println!("No results");
        return;
    }
    let headers = ["id", "name", "status", "type", "region", "created_at"];
    println!("{}", headers.join("\t"));
    for item in items {
        let mut row = Vec::new();
        for h in headers {
            row.push(truncate(
                item.get(h).and_then(Value::as_str).unwrap_or(""),
                full,
            ));
        }
        println!("{}", row.join("\t"));
    }
}

fn print_object(map: &serde_json::Map<String, Value>, full: bool) {
    for (k, v) in map {
        println!("{k}\t{}", truncate(&scalar(v), full));
    }
}

fn scalar(v: &Value) -> String {
    match v {
        Value::Null => "".into(),
        Value::String(s) => s.clone(),
        Value::Bool(b) => b.to_string(),
        Value::Number(n) => n.to_string(),
        _ => serde_json::to_string(v).unwrap_or_default(),
    }
}

fn truncate(s: &str, full: bool) -> String {
    if full || s.chars().count() <= 80 {
        s.to_string()
    } else {
        format!("{}…", s.chars().take(79).collect::<String>())
    }
}

fn encode(s: &str) -> String {
    url::form_urlencoded::byte_serialize(s.as_bytes()).collect()
}
fn limit() -> (&'static str, String) {
    ("limit", "100".into())
}

async fn resolve(api: &ApiClient, resource: &str, name: &str) -> Result<String> {
    resolve_unscoped(api, resource, name).await
}

async fn resolve_unscoped(api: &ApiClient, resource: &str, name: &str) -> Result<String> {
    resolve_scoped_with_id(api, resource, name, None).await
}

async fn resolve_scoped(
    api: &ApiClient,
    resource: &str,
    name: &str,
    scope: Option<(&str, &str)>,
) -> Result<String> {
    let scope_id = if let Some((value, key)) = scope {
        let resource_name = match key.trim_end_matches("_id") {
            "app" => "apps",
            "service" => "services",
            "deployment" => "deployments",
            "regional_deployment" => "regional_deployments",
            other => other,
        };
        Some((
            if looks_like_id(value) {
                value.to_string()
            } else {
                resolve_unscoped(api, resource_name, value).await?
            },
            key,
        ))
    } else {
        None
    };
    resolve_scoped_with_id(api, resource, name, scope_id).await
}

async fn resolve_scoped_with_id(
    api: &ApiClient,
    resource: &str,
    name: &str,
    scope: Option<(String, &str)>,
) -> Result<String> {
    if looks_like_id(name) {
        return Ok(name.to_string());
    }
    let mut q = vec![("name", name.to_string()), limit()];
    if let Some((value, key)) = scope {
        q.push((key, value));
    }
    let path = match resource {
        "persistent_volumes" => "/v1/persistent_volumes".to_string(),
        _ => format!("/v1/{resource}"),
    };
    let value = api.get(&path, &q).await?;
    let arr = value
        .get(resource)
        .and_then(Value::as_array)
        .or_else(|| {
            value
                .get(resource.replace('_', "-"))
                .and_then(Value::as_array)
        })
        .or_else(|| value.as_array())
        .ok_or_else(|| anyhow!("unable to resolve {resource} `{name}`"))?;
    let matches: Vec<_> = arr
        .iter()
        .filter(|v| {
            v.get("name").and_then(Value::as_str) == Some(name)
                || v.get("id").and_then(Value::as_str) == Some(name)
        })
        .collect();
    match matches.len() {
        1 => Ok(matches[0]
            .get("id")
            .and_then(Value::as_str)
            .unwrap_or(name)
            .to_string()),
        0 => bail!("{resource} `{name}` not found"),
        _ => bail!("{resource} `{name}` is ambiguous; use its ID"),
    }
}

fn looks_like_id(s: &str) -> bool {
    s.len() >= 8
        && s.chars()
            .all(|c| c.is_ascii_alphanumeric() || c == '-' || c == '_')
        && s.chars().any(|c| c.is_ascii_digit())
}

async fn ensure_app(api: &ApiClient, name: &str) -> Result<String> {
    match resolve(api, "apps", name).await {
        Ok(id) => Ok(id),
        Err(_) => {
            let value = api.post("/v1/apps", json!({"app": {"name": name}})).await?;
            value
                .pointer("/app/id")
                .and_then(Value::as_str)
                .map(str::to_string)
                .ok_or_else(|| anyhow!("app creation response did not contain an id"))
        }
    }
}

fn split_app_service(app: Option<&str>, service: &str) -> Result<(String, String)> {
    if let Some((a, s)) = service.split_once('/') {
        return Ok((a.to_string(), s.to_string()));
    }
    Ok((
        app.ok_or_else(|| anyhow!("service must be APP/SERVICE or --app must be provided"))?
            .to_string(),
        service.to_string(),
    ))
}

fn service_body(
    name: &str,
    args: ServiceDefinitionArgs,
    app_id: Option<String>,
    override_: bool,
) -> Result<Value> {
    let mut service = json!({"name": name});
    if let Some(app_id) = app_id {
        service["app_id"] = json!(app_id);
    }
    if let Some(v) = args.instance_type {
        service["instance_type"] = json!(v);
    }
    if !args.regions.is_empty() {
        service["regions"] = json!(args.regions);
    }
    if args.privileged {
        service["privileged"] = json!(true);
    }
    if !args.env.is_empty() || !args.env_file.is_empty() {
        service["env"] = json!(parse_env(args.env, args.env_file)?);
    }
    if !args.ports.is_empty() {
        service["ports"] = json!(
            args.ports
                .iter()
                .map(|p| parse_colon_map(p))
                .collect::<Vec<_>>()
        );
    }
    if !args.routes.is_empty() {
        service["routes"] = json!(
            args.routes
                .iter()
                .map(|p| parse_colon_map(p))
                .collect::<Vec<_>>()
        );
    }
    if !args.volumes.is_empty() {
        service["volumes"] = json!(
            args.volumes
                .iter()
                .map(|p| parse_colon_map(p))
                .collect::<Vec<_>>()
        );
    }
    if !args.config_file.is_empty() {
        service["config_files"] = json!(args.config_file);
    }
    if !args.secrets.is_empty() {
        service["secrets"] = json!(args.secrets);
    }
    if !args.checks.is_empty() {
        service["health_checks"] = json!(args.checks);
    }
    if !args.checks_grace_period.is_empty() {
        service["checks_grace_period"] = json!(args.checks_grace_period);
    }
    if !args.proxy_ports.is_empty() {
        service["proxy_ports"] = json!(args.proxy_ports);
    }
    if !args.scale.is_empty() {
        service["scaling"] = scaling_body(args.scale, 1, vec![]);
    }
    if let Some(v) = args.min_scale {
        service["min_scale"] = json!(v);
    }
    if let Some(v) = args.max_scale {
        service["max_scale"] = json!(v);
    }
    if let Some(v) = args.autoscaling_target {
        service["autoscaling_target"] = json!(v);
    }
    if args.light_sleep_delay.0.as_secs() > 0 {
        service["light_sleep_delay"] = json!(args.light_sleep_delay.0.as_secs());
    }
    if args.deep_sleep_delay.0.as_secs() > 0 {
        service["deep_sleep_delay"] = json!(args.deep_sleep_delay.0.as_secs());
    }
    if args.delete_after_delay.0.as_secs() > 0 {
        service["delete_after_delay"] = json!(args.delete_after_delay.0.as_secs());
    }
    if args.delete_after_inactivity_delay.0.as_secs() > 0 {
        service["delete_after_inactivity_delay"] =
            json!(args.delete_after_inactivity_delay.0.as_secs());
    }
    let mut definition = json!({});
    if let Some(git) = args.git {
        definition["git"] = json!({"repository": git, "branch": args.git_branch, "sha": args.git_sha, "builder": args.git_builder, "build_command": args.git_build_command, "run_command": args.git_run_command, "privileged": args.git_privileged, "docker": compact(json!({"dockerfile": args.git_docker_dockerfile, "entrypoint": args.git_docker_entrypoint, "args": args.git_docker_args}))});
    }
    if let Some(docker) = args.docker {
        definition["docker"] = json!({"image": docker, "entrypoint": args.docker_entrypoint, "args": args.docker_args});
    }
    if let Some(archive) = args.archive {
        definition["archive"] = json!({"id": archive, "builder": args.archive_builder, "build_command": args.archive_build_command, "run_command": args.archive_run_command, "docker": compact(json!({"dockerfile": args.archive_docker_dockerfile, "entrypoint": args.archive_docker_entrypoint, "args": args.archive_docker_args}))});
    }
    service["definition"] = compact(definition);
    Ok(json!({"service": compact(service), "override": override_}))
}

fn compact(mut v: Value) -> Value {
    match &mut v {
        Value::Object(map) => {
            let keys: Vec<_> = map
                .iter()
                .filter(|(_, v)| {
                    v.is_null()
                        || v.as_array().is_some_and(Vec::is_empty)
                        || v.as_object().is_some_and(serde_json::Map::is_empty)
                })
                .map(|(k, _)| k.clone())
                .collect();
            for k in keys {
                map.remove(&k);
            }
            for value in map.values_mut() {
                *value = compact(value.take());
            }
        }
        Value::Array(arr) => {
            for value in arr {
                *value = compact(value.take());
            }
        }
        _ => {}
    }
    v
}

fn parse_key_values(values: Vec<String>) -> Result<BTreeMap<String, String>> {
    let mut map = BTreeMap::new();
    for item in values {
        let (k, v) = item
            .split_once('=')
            .ok_or_else(|| anyhow!("expected KEY=VALUE, got {item}"))?;
        map.insert(k.to_string(), v.to_string());
    }
    Ok(map)
}

fn parse_env(values: Vec<String>, files: Vec<PathBuf>) -> Result<Vec<Value>> {
    let mut map = parse_key_values(values)?;
    for file in files {
        for line in fs::read_to_string(file)?.lines() {
            let line = line.trim();
            if line.is_empty() || line.starts_with('#') {
                continue;
            }
            let (k, v) = line
                .split_once('=')
                .ok_or_else(|| anyhow!("expected KEY=VALUE in env file"))?;
            map.insert(k.to_string(), v.to_string());
        }
    }
    Ok(map
        .into_iter()
        .map(|(k, v)| json!({"key": k, "value": v}))
        .collect())
}

fn parse_colon_map(value: &str) -> Value {
    if value.contains('=') {
        json!(
            value
                .split(',')
                .filter_map(|p| p.split_once('='))
                .map(|(k, v)| (k.to_string(), Value::String(v.to_string())))
                .collect::<serde_json::Map<_, _>>()
        )
    } else if value.contains(':') {
        let parts: Vec<_> = value.split(':').collect();
        json!(parts)
    } else {
        json!(value)
    }
}

fn scaling_body(scale: Vec<String>, instances: i64, regions: Vec<String>) -> Value {
    let mut regional = Vec::new();
    for s in scale {
        if let Some(region) = s.strip_prefix('!') {
            regional.push(json!({"region": region, "instances": null}));
        } else if let Some((region, count)) = s.split_once(':') {
            regional.push(
                json!({"region": region, "instances": count.parse::<i64>().unwrap_or(instances)}),
            );
        }
    }
    for region in regions {
        if let Some(r) = region.strip_prefix('!') {
            regional.push(json!({"region": r, "instances": null}));
        } else {
            regional.push(json!({"region": region, "instances": instances}));
        }
    }
    if regional.is_empty() {
        json!({"instances": instances})
    } else {
        json!({"regions": regional})
    }
}

fn database_body(
    name: String,
    app_id: Option<String>,
    pg_version: Option<i64>,
    region: Option<String>,
    instance_type: Option<String>,
    db_owner: Option<String>,
    db_name: Option<String>,
) -> Value {
    compact(
        json!({"service": {"name": name, "app_id": app_id, "type": "DATABASE", "database": {"engine": "postgresql", "version": pg_version, "owner": db_owner, "name": db_name}, "regions": region.map(|r| vec![r]), "instance_type": instance_type}}),
    )
}

fn sandbox_body(args: SandboxCreateArgs) -> Result<Value> {
    let mut def = ServiceDefinitionArgs {
        docker: args.docker,
        docker_entrypoint: args.docker_entrypoint,
        docker_args: args.docker_args,
        regions: args.regions,
        env: args.env,
        config_file: args.config_file,
        delete_after_delay: args.delete_after_delay,
        delete_after_inactivity_delay: args.delete_after_inactivity_delay,
        light_sleep_delay: args.light_sleep_delay,
        deep_sleep_delay: args.deep_sleep_delay,
        ..Default::default()
    };
    let _ = args.wait_timeout;
    if def.docker.is_none() {
        def.docker = Some("ubuntu:latest".into());
    }
    let mut body = service_body(&args.name, def, None, false)?;
    body["service"]["type"] = json!("SANDBOX");
    Ok(body)
}

fn read_secret_value(value: Option<String>, filename: Option<PathBuf>) -> Result<String> {
    match (value, filename) {
        (Some(v), None) => Ok(v),
        (_, Some(path)) => Ok(fs::read_to_string(path)?.trim_end().to_string()),
        (None, None) => {
            let mut s = String::new();
            io::stdin().read_to_string(&mut s)?;
            Ok(s.trim_end().to_string())
        }
    }
}

fn read_yaml(path: &Path) -> Result<Value> {
    Ok(serde_yaml::from_str(&fs::read_to_string(path)?)?)
}

fn create_tarball(path: &Path, ignore_dirs: &[String]) -> Result<Vec<u8>> {
    let enc = flate2::write::GzEncoder::new(Vec::new(), flate2::Compression::default());
    let mut tar = tar::Builder::new(enc);
    if path.is_file() {
        tar.append_path(path)?;
    } else {
        append_dir(&mut tar, path, path, ignore_dirs)?;
    }
    let enc = tar.into_inner()?;
    Ok(enc.finish()?)
}

fn append_dir<W: Write>(
    tar: &mut tar::Builder<W>,
    root: &Path,
    dir: &Path,
    ignore_dirs: &[String],
) -> Result<()> {
    for entry in fs::read_dir(dir)? {
        let entry = entry?;
        let path = entry.path();
        let name = entry.file_name().to_string_lossy().to_string();
        if path.is_dir() && ignore_dirs.iter().any(|i| i == &name) {
            continue;
        }
        let rel = path.strip_prefix(root)?;
        if path.is_dir() {
            tar.append_dir(rel, &path)?;
            append_dir(tar, root, &path, ignore_dirs)?;
        } else {
            tar.append_path_with_name(&path, rel)?;
        }
    }
    Ok(())
}

async fn wait_for_service(api: &ApiClient, initial: &Value, timeout: Duration) -> Result<()> {
    let id = initial
        .pointer("/service/id")
        .and_then(Value::as_str)
        .or_else(|| {
            initial
                .pointer("/deployment/service_id")
                .and_then(Value::as_str)
        })
        .unwrap_or_default();
    if id.is_empty() || timeout.is_zero() {
        return Ok(());
    }
    let start = std::time::Instant::now();
    while start.elapsed() < timeout {
        let value = api
            .get(&format!("/v1/services/{}", encode(id)), &[])
            .await?;
        let status = value
            .pointer("/service/status")
            .and_then(Value::as_str)
            .unwrap_or_default();
        if matches!(status, "HEALTHY" | "RUNNING" | "READY") {
            return Ok(());
        }
        tokio::time::sleep(Duration::from_secs(5)).await;
    }
    bail!("timed out waiting for service deployment")
}

async fn service_logs(cfg: &Config, api: &ApiClient, args: LogsArgs) -> Result<()> {
    let service_id = resolve_scoped(
        api,
        "services",
        &args.name,
        args.app.as_deref().map(|a| (a, "app_id")),
    )
    .await?;
    logs_query(
        cfg,
        api,
        "service_id",
        &service_id,
        args.instance,
        args.log_type,
        args.tail,
        args.start_time.or(args.since),
        args.end_time,
        args.regex_search,
        args.text_search,
        args.order,
    )
    .await
}

async fn logs_query(
    cfg: &Config,
    api: &ApiClient,
    key: &str,
    id: &str,
    instance: Option<String>,
    log_type: Option<String>,
    tail: bool,
    start: Option<String>,
    end: Option<String>,
    regex: Option<String>,
    text: Option<String>,
    order: String,
) -> Result<()> {
    let mut q = vec![(key, id.to_string()), ("order", order)];
    if let Some(v) = instance {
        q.push(("instance_id", v));
    }
    if let Some(v) = log_type {
        q.push(("type", v));
    }
    if tail {
        q.push(("tail", "true".into()));
    }
    if let Some(v) = start {
        q.push(("start_time", v));
    }
    if let Some(v) = end {
        q.push(("end_time", v));
    }
    if let Some(v) = regex {
        q.push(("regex_search", v));
    }
    if let Some(v) = text {
        q.push(("text_search", v));
    }
    render(cfg, api.get("/v1/logs", &q).await?)
}

async fn exec_remote(
    api: &ApiClient,
    cfg: &Config,
    resource: &str,
    name: &str,
    app: Option<&str>,
    cmd: String,
    args: Vec<String>,
) -> Result<()> {
    let id = if resource == "services" {
        resolve_scoped(api, resource, name, app.map(|a| (a, "app_id"))).await?
    } else {
        resolve(api, resource, name).await?
    };
    render(
        cfg,
        api.post(
            &format!("/v1/{resource}/{}/exec", encode(&id)),
            json!({"command": cmd, "args": args}),
        )
        .await?,
    )
}

async fn cp_remote(
    api: &ApiClient,
    cfg: &Config,
    source: String,
    destination: String,
) -> Result<()> {
    let (instance, remote, upload) = if let Some((inst, path)) = source.split_once(':') {
        (inst.to_string(), path.to_string(), false)
    } else if let Some((inst, path)) = destination.split_once(':') {
        (inst.to_string(), path.to_string(), true)
    } else {
        bail!("one path must be remote in INSTANCE:PATH format");
    };
    let id = resolve(api, "instances", &instance).await?;
    if upload {
        let data = base64::Engine::encode(
            &base64::engine::general_purpose::STANDARD,
            fs::read(source)?,
        );
        render(
            cfg,
            api.post(
                &format!("/v1/instances/{}/fs/upload", encode(&id)),
                json!({"path": remote, "data": data}),
            )
            .await?,
        )
    } else {
        let res = api
            .post(
                &format!("/v1/instances/{}/fs/download", encode(&id)),
                json!({"path": remote}),
            )
            .await?;
        if let Some(data) = res.get("data").and_then(Value::as_str) {
            fs::write(
                destination,
                base64::Engine::decode(&base64::engine::general_purpose::STANDARD, data)?,
            )?;
            Ok(())
        } else {
            render(cfg, res)
        }
    }
}

async fn sandbox_rpc(
    cfg: &Config,
    api: &ApiClient,
    name: &str,
    action: &str,
    body: Value,
) -> Result<()> {
    render(cfg, sandbox_rpc_value(api, name, action, body).await?)
}
async fn sandbox_rpc_value(
    api: &ApiClient,
    name: &str,
    action: &str,
    body: Value,
) -> Result<Value> {
    let id = resolve(api, "services", name).await?;
    match api
        .post(
            &format!("/v1/sandboxes/{}/{}", encode(&id), action),
            body.clone(),
        )
        .await
    {
        Ok(value) => Ok(value),
        Err(_) => {
            api.post(
                &format!("/v1/services/{}/sandbox/{}", encode(&id), action),
                body,
            )
            .await
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use clap::CommandFactory;

    #[test]
    fn cli_builds() {
        Cli::command().debug_assert();
    }

    #[test]
    fn parses_duration_units() {
        assert_eq!(parse_duration("5m").unwrap().as_secs(), 300);
        assert_eq!(parse_duration("1h30m").unwrap().as_secs(), 5400);
        assert_eq!(parse_duration("42").unwrap().as_secs(), 42);
    }

    #[test]
    fn parses_key_values() {
        let map = parse_key_values(vec!["A=B".into(), "C=D".into()]).unwrap();
        assert_eq!(map.get("A").unwrap(), "B");
        assert!(parse_key_values(vec!["bad".into()]).is_err());
    }

    #[test]
    fn builds_scaling_body() {
        assert_eq!(
            scaling_body(vec!["fra:3".into()], 1, vec![]),
            json!({"regions": [{"region": "fra", "instances": 3}]})
        );
        assert_eq!(
            scaling_body(vec![], 2, vec!["was".into()]),
            json!({"regions": [{"region": "was", "instances": 2}]})
        );
    }

    #[test]
    fn splits_app_service() {
        assert_eq!(
            split_app_service(None, "app/svc").unwrap(),
            ("app".into(), "svc".into())
        );
        assert!(split_app_service(None, "svc").is_err());
    }
}
