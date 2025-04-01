package mitigator

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
)

const (
	RequestWisdomCommand = "request_wisdom"
	Delimiter            = '\n'
)

// HandleConnection implements the tcp.HandlerFunc interface.
// It reads requests, gets wisdom, and sends responses.
func (h *Controller) HandleConnection(ctx context.Context, conn net.Conn) error {
	clientAddr := conn.RemoteAddr().String()
	h.log.InfoContext(ctx, "Handling new connection", slog.String("client_addr", clientAddr))

	// Set a deadline for the entire handling process for this connection
	handlerCtx, cancel := context.WithTimeout(ctx, h.handlerTimeout)
	defer cancel() // Ensure resources are released

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// 1. Read the request command
	// The client sends "request_wisdom\n" after PoW
	commandBytes, err := reader.ReadBytes(Delimiter)
	if err != nil {
		if errors.Is(err, io.EOF) {
			h.log.WarnContext(handlerCtx, "Client closed connection before sending command", slog.String("client_addr", clientAddr))
			return nil // Not necessarily an error, client just disconnected
		}
		h.log.ErrorContext(handlerCtx, "Failed to read command", slog.String("error", err.Error()), slog.String("client_addr", clientAddr))
		return fmt.Errorf("reading command: %w", err)
	}

	command := string(commandBytes[:len(commandBytes)-1]) // Remove delimiter
	h.log.DebugContext(handlerCtx, "Received command", slog.String("command", command), slog.String("client_addr", clientAddr))

	// 2. Validate command (optional but good practice)
	if command != RequestWisdomCommand {
		h.log.WarnContext(handlerCtx, "Received unknown command", slog.String("command", command), slog.String("client_addr", clientAddr))
		// Optionally send an error response before closing
		_, wErr := writer.WriteString(fmt.Sprintf("error: unknown command '%s'%c", command, Delimiter))
		if wErr == nil {
			wErr = writer.Flush()
		}
		if wErr != nil {
			h.log.ErrorContext(handlerCtx, "Failed to write unknown command error response", slog.String("error", wErr.Error()), slog.String("client_addr", clientAddr))
		}
		return fmt.Errorf("unknown command: %s", command) // Close connection after unknown command
	}

	// 3. Get Wisdom using the policy
	wisdom, err := h.wisdomProvider.GetWisdom(handlerCtx)
	if err != nil {
		h.log.ErrorContext(handlerCtx, "Failed to get wisdom", slog.String("error", err.Error()), slog.String("client_addr", clientAddr))
		// Send an error response back to the client
		_, wErr := writer.WriteString(fmt.Sprintf("error: %s%c", err.Error(), Delimiter))
		if wErr == nil {
			wErr = writer.Flush()
		}
		if wErr != nil {
			h.log.ErrorContext(handlerCtx, "Failed to write wisdom error response", slog.String("error", wErr.Error()), slog.String("client_addr", clientAddr))
		}
		return fmt.Errorf("getting wisdom: %w", err)
	}

	// 4. Marshal the response (assuming JSON, though client reads raw string)
	// Let's keep it simple and send the quote directly as requested by client's current logic.
	// If JSON is needed later:
	// responseBytes, err := json.Marshal(wisdom)
	// if err != nil {
	//  h.log.ErrorContext(handlerCtx, "Failed to marshal wisdom response", slog.String("error", err.Error()), slog.String("client_addr", clientAddr))
	//  return fmt.Errorf("marshaling response: %w", err)
	// }
	responseString := wisdom.Quote

	// 5. Send the response back to the client
	_, err = writer.WriteString(responseString + string(Delimiter))
	if err != nil {
		h.log.ErrorContext(handlerCtx, "Failed to write wisdom response", slog.String("error", err.Error()), slog.String("client_addr", clientAddr))
		return fmt.Errorf("writing response: %w", err)
	}

	// 6. Flush the writer buffer
	err = writer.Flush()
	if err != nil {
		h.log.ErrorContext(handlerCtx, "Failed to flush writer", slog.String("error", err.Error()), slog.String("client_addr", clientAddr))
		return fmt.Errorf("flushing writer: %w", err)
	}

	h.log.InfoContext(handlerCtx, "Successfully handled connection", slog.String("client_addr", clientAddr), slog.String("wisdom_sent", responseString))

	// Connection will be closed by the tcp.Server when this handler returns
	return nil
}

// Ensure Handler implements the required interface from tcp package
var _ tcp.HandlerFunc = (&Handler{}).HandleConnection
