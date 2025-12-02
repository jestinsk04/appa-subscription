package services

import (
	"appa_subscriptions/internal/models"
	helpers "appa_subscriptions/pkg"
	"appa_subscriptions/pkg/mailgun"
	"context"

	"go.uber.org/zap"
)

type notificationService struct {
	muRepo mailgun.Repository
	logger *zap.Logger
}

type notificationJob struct {
	vars     models.ConfirmationOrderEmailVars
	email    string
	template string
}

var notificationJobsQueue chan notificationJob

var (
	EmailsTemplates = map[string]mailgun.SendEmailRequest{
		"create_order": {
			Subject:  "💙 Mantén tu cobertura Appa activa",
			Template: "Cuota creada",
		},
		"reminder": {
			Subject:  "🐾 Tu cobertura Appa sigue pendiente de pago",
			Template: "Recordatorio dia 1 - 7 - 14 -27",
		},
		"cancellation": {
			Subject:  "🚫 Tu póliza ha sido cancelada (puedes reactivarla)",
			Template: "Cancelación",
		},
		"reactivation": {
			Subject:  "💙 Vuelve a estar protegido con Appa (preexistencias aplican)",
			Template: "Cancelación",
		},
	}
)

func NewNotificationService(
	mailgunRepo mailgun.Repository,
	logger *zap.Logger,
) *notificationService {

	service := &notificationService{
		muRepo: mailgunRepo,
		logger: logger,
	}

	service.startWorkerPool(5)

	return service
}

func (h *notificationService) startWorkerPool(numWorkers int) {
	notificationJobsQueue = make(chan notificationJob, 100) // Cola con buffer de 100 trabajos

	for i := 1; i <= numWorkers; i++ {
		go h.worker(i, notificationJobsQueue)
	}
	h.logger.Info("Started workers", zap.Int("num_workers", numWorkers))
}

func (h *notificationService) worker(id int, jobs <-chan notificationJob) {
	for job := range jobs {
		err := h.sendEmail(context.Background(), job.vars, job.email, job.template)
		if err != nil {
			h.logger.Error("error sending email", zap.Error(err), zap.Int("worker_id", id), zap.String("to", job.email), zap.String("template", job.template))
		}
		h.logger.Info("Worker finished job for order", zap.String("template", job.template))
	}
}

func (s *notificationService) sendEmail(
	ctx context.Context,
	vars models.ConfirmationOrderEmailVars,
	email string,
	template string,
) error {
	varsForEmail := helpers.GetVarsForConfirmationOrderEmail(vars)

	emailTemplate := EmailsTemplates[template]
	emailTemplate.Vars = varsForEmail
	emailTemplate.To = email

	err := s.muRepo.SendEmail(ctx, emailTemplate)
	if err != nil {
		return err
	}

	return nil
}
