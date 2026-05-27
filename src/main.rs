fn main() {
    if let Err(err) = koyeb_cli::run_blocking() {
        eprintln!("{err}");
        std::process::exit(1);
    }
}
