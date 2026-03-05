package services

import (
	"context"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
)

// NotificationService maneja la lógica de notificaciones in-app
type NotificationService interface {
	// Notifica a todos los usuarios de un apartamento que llegó un paquete
	NotifyPackage(ctx context.Context, condominioID, apartmentID, packageID, carrier, apartmentLabel string) error

	// Notifica al dueño de una reserva que cambió su estado
	NotifyReservationStatus(ctx context.Context, condominioID, userID, reservationID, espacioNombre, newEstado string) error

	// Lista notificaciones del usuario actual
	List(ctx context.Context, userID, condominioID string, page, pageSize int) ([]*models.Notification, int, error)

	// Marca una notificación como leída
	MarkRead(ctx context.Context, notifID, userID string) error

	// Marca todas las notificaciones como leídas
	MarkAllRead(ctx context.Context, userID, condominioID string) error

	// Cantidad de notificaciones sin leer
	UnreadCount(ctx context.Context, userID, condominioID string) (int, error)

	// Elimina una notificación
	Delete(ctx context.Context, notifID, userID string) error
}

type notificationServiceImpl struct {
	repo     repositories.NotificationRepository
	userRepo repositories.UserRepository
}

func NewNotificationService(repo repositories.NotificationRepository, userRepo repositories.UserRepository) NotificationService {
	return &notificationServiceImpl{repo: repo, userRepo: userRepo}
}

func (s *notificationServiceImpl) NotifyPackage(ctx context.Context, condominioID, apartmentID, packageID, carrier, apartmentLabel string) error {
	users, err := s.userRepo.GetByApartment(ctx, apartmentID)
	if err != nil || len(users) == 0 {
		return err
	}

	refID := packageID
	for _, u := range users {
		notif := &models.Notification{
			UserID:       u.ID,
			CondominioID: condominioID,
			Tipo:         "paquete",
			Titulo:       "Paquete recibido",
			Mensaje:      "Llegó un paquete de " + carrier + " para " + apartmentLabel + ". Puedes retirarlo en la portería.",
			ReferenciaID: &refID,
		}
		_ = s.repo.Create(ctx, notif) // best-effort: no falla la operación principal
	}
	return nil
}

func (s *notificationServiceImpl) NotifyReservationStatus(ctx context.Context, condominioID, userID, reservationID, espacioNombre, newEstado string) error {
	var titulo, mensaje string
	switch newEstado {
	case "confirmada":
		titulo = "Reserva confirmada"
		mensaje = "Tu reserva de \"" + espacioNombre + "\" fue confirmada."
	case "cancelada":
		titulo = "Reserva cancelada"
		mensaje = "Tu reserva de \"" + espacioNombre + "\" fue cancelada."
	case "rechazada":
		titulo = "Reserva rechazada"
		mensaje = "Tu reserva de \"" + espacioNombre + "\" fue rechazada por la administración."
	case "pendiente":
		titulo = "Reserva en revisión"
		mensaje = "Tu reserva de \"" + espacioNombre + "\" está pendiente de aprobación."
	default:
		titulo = "Actualización de reserva"
		mensaje = "El estado de tu reserva de \"" + espacioNombre + "\" cambió a " + newEstado + "."
	}

	refID := reservationID
	notif := &models.Notification{
		UserID:       userID,
		CondominioID: condominioID,
		Tipo:         "reserva",
		Titulo:       titulo,
		Mensaje:      mensaje,
		ReferenciaID: &refID,
	}
	return s.repo.Create(ctx, notif)
}

func (s *notificationServiceImpl) List(ctx context.Context, userID, condominioID string, page, pageSize int) ([]*models.Notification, int, error) {
	return s.repo.ListByUser(ctx, userID, condominioID, page, pageSize)
}

func (s *notificationServiceImpl) MarkRead(ctx context.Context, notifID, userID string) error {
	return s.repo.MarkRead(ctx, notifID, userID)
}

func (s *notificationServiceImpl) MarkAllRead(ctx context.Context, userID, condominioID string) error {
	return s.repo.MarkAllRead(ctx, userID, condominioID)
}

func (s *notificationServiceImpl) UnreadCount(ctx context.Context, userID, condominioID string) (int, error) {
	return s.repo.UnreadCount(ctx, userID, condominioID)
}

func (s *notificationServiceImpl) Delete(ctx context.Context, notifID, userID string) error {
	return s.repo.Delete(ctx, notifID, userID)
}
