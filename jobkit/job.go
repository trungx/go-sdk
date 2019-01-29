package jobkit

import (
	"context"
	"fmt"
	"time"

	"github.com/blend/go-sdk/cron"
	"github.com/blend/go-sdk/diagnostics"
	"github.com/blend/go-sdk/email"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/slack"
	"github.com/blend/go-sdk/stats"
	"github.com/blend/go-sdk/uuid"
)

var (
	_ cron.Job                    = (*Job)(nil)
	_ cron.OnStartReceiver        = (*Job)(nil)
	_ cron.OnCompleteReceiver     = (*Job)(nil)
	_ cron.OnFailureReceiver      = (*Job)(nil)
	_ cron.OnCancellationReceiver = (*Job)(nil)
	_ cron.OnBrokenReceiver       = (*Job)(nil)
	_ cron.OnFixedReceiver        = (*Job)(nil)
)

// NewJob creates a new exec job.
func NewJob(action func(context.Context) error) *Job {
	return &Job{
		name:   uuid.V4().String(),
		action: action,
	}
}

// Job is the main job body.
type Job struct {
	name          string
	notifications *NotificationsConfig

	schedule cron.Schedule
	timeout  time.Duration
	action   func(context.Context) error

	log         logger.Log
	statsClient stats.Collector
	slackClient slack.Sender
	emailClient email.Sender
	errorClient diagnostics.Notifier
}

// Name returns the job name.
func (job Job) Name() string {
	return job.name
}

// WithName sets the name.
func (job *Job) WithName(name string) *Job {
	job.name = name
	return job
}

// Schedule returns the job schedule.
func (job Job) Schedule() cron.Schedule {
	return job.schedule
}

// WithSchedule sets the schedule.
func (job *Job) WithSchedule(schedule cron.Schedule) *Job {
	job.schedule = schedule
	return job
}

// NotificationsConfig returns the managment server config.
func (job Job) NotificationsConfig() *NotificationsConfig {
	return job.notifications
}

// WithNotificationsConfig sets the config.
func (job *Job) WithNotificationsConfig(cfg *NotificationsConfig) *Job {
	job.notifications = cfg
	return job
}

// Timeout returns the timeout.
func (job Job) Timeout() time.Duration {
	return job.timeout
}

// WithTimeout sets the job timeout.
func (job *Job) WithTimeout(d time.Duration) *Job {
	job.timeout = d
	return job
}

// WithLogger sets the job logger.
func (job *Job) WithLogger(log logger.Log) *Job {
	job.log = log
	return job
}

// WithStatsClient sets the job stats client.
func (job *Job) WithStatsClient(client stats.Collector) *Job {
	job.statsClient = client
	return job
}

// WithSlackClient sets the job slack client.
func (job *Job) WithSlackClient(client slack.Sender) *Job {
	job.slackClient = client
	return job
}

// WithEmailClient sets the job email client.
func (job *Job) WithEmailClient(client email.Sender) *Job {
	job.emailClient = client
	return job
}

// WithErrorClient sets the job error client.
func (job *Job) WithErrorClient(client diagnostics.Notifier) *Job {
	job.errorClient = client
	return job
}

// OnStart is a lifecycle event handler.
func (job Job) OnStart(ctx context.Context) {
	if job.notifications != nil && job.notifications.NotifyOnStartOrDefault() {
		job.notify(ctx, cron.FlagStarted)
	}
}

// OnComplete is a lifecycle event handler.
func (job Job) OnComplete(ctx context.Context) {
	if job.notifications != nil && job.notifications.NotifyOnSuccessOrDefault() {
		job.notify(ctx, cron.FlagComplete)
	}
}

// OnFailure is a lifecycle event handler.
func (job Job) OnFailure(ctx context.Context) {
	if job.notifications != nil && job.notifications.NotifyOnFailureOrDefault() {
		job.notify(ctx, cron.FlagFailed)
	}
}

// OnBroken is a lifecycle event handler.
func (job Job) OnBroken(ctx context.Context) {
	if job.notifications != nil && job.notifications.NotifyOnBrokenOrDefault() {
		job.notify(ctx, cron.FlagBroken)
	}
}

// OnFixed is a lifecycle event handler.
func (job Job) OnFixed(ctx context.Context) {
	if job.notifications != nil && job.notifications.NotifyOnFixedOrDefault() {
		job.notify(ctx, cron.FlagFixed)
	}
}

// OnCancellation is a lifecycle event handler.
func (job Job) OnCancellation(ctx context.Context) {
	if job.notifications != nil && job.notifications.NotifyOnFailureOrDefault() {
		job.notify(ctx, cron.FlagCancelled)
	}
}

func (job Job) notify(ctx context.Context, flag logger.Flag) {
	if job.statsClient != nil {
		job.statsClient.Increment(string(flag), fmt.Sprintf("%s:%s", stats.TagJob, job.Name()))
		if ji := cron.GetJobInvocation(ctx); ji != nil {
			logger.MaybeError(job.log, job.statsClient.TimeInMilliseconds(string(flag), ji.Elapsed, fmt.Sprintf("%s:%s", stats.TagJob, job.Name())))
		}
	}
	if job.slackClient != nil {
		if ji := cron.GetJobInvocation(ctx); ji != nil {
			logger.MaybeError(job.log, job.slackClient.Send(context.Background(), NewSlackMessage(string(flag), ji)))
		}
	}
	if job.emailClient != nil {
		if ji := cron.GetJobInvocation(ctx); ji != nil {
			message, err := NewEmailMessage(string(flag), ji)
			if err != nil {
				logger.MaybeError(job.log, err)
			}
			logger.MaybeError(job.log, job.emailClient.Send(context.Background(), message))
		}
	}
	if job.errorClient != nil {
		if ji := cron.GetJobInvocation(ctx); ji != nil && ji.Err != nil {
			job.errorClient.Notify(ji.Err)
		}
	}
}

// Execute is the job body.
func (job Job) Execute(ctx context.Context) error {
	return job.action(ctx)
}
