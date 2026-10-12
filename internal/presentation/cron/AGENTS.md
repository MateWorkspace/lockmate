# Cron presentation

Schedule internal maintenance through domain usecase interfaces. This is a
delivery entrypoint like HTTP; it holds no business rules. Composition wires the
cron runner inside the presentation layer together with the HTTP route.

- `runner/runner.go` is the scheduling analog of `route.Route`. `Run(ctx, Args)`
  runs every job once synchronously (a required failure aborts startup; optional
  jobs never do), then schedules each job with the `github.com/robfig/cron/v3`
  engine using an `@every <interval>` spec and a bounded per-run timeout. It
  returns a stop function that stops the scheduler and waits for running jobs;
  composition calls it during shutdown before PostgreSQL closes.
- `job/` holds one file per job, each exposing a function that adapts a domain
  usecase to a `func(ctx) error`. The signing-key job invokes
  Misc.SigningKey.Reconcile; the auth-flow job invokes Misc.AuthFlow.PurgeExpired.
  Keep signing-key reconciliation required and best-effort cleanup optional.

Keep jobs idempotent and cancellation-aware. Jobs log their own failures; a
required startup failure is also returned so launch can abort. Do not reimplement
scheduling, add logging wrappers, or perform business validation here. Tests live
under `tests/presentation/cron`.
