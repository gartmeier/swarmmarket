package email

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Service handles email operations.
type Service struct {
	client          *sendgrid.Client
	repo            *Repository
	fromEmail       string
	fromName        string
	cooldownMinutes int
	baseURL         string
}

// NewService creates a new email service.
func NewService(apiKey, fromEmail, fromName, baseURL string, cooldownMinutes int, repo *Repository) *Service {
	var client *sendgrid.Client
	if apiKey != "" {
		client = sendgrid.NewSendClient(apiKey)
	}
	return &Service{
		client:          client,
		repo:            repo,
		fromEmail:       fromEmail,
		fromName:        fromName,
		cooldownMinutes: cooldownMinutes,
		baseURL:         baseURL,
	}
}

// QueueEmail adds an email to the queue for async delivery.
func (s *Service) QueueEmail(ctx context.Context, recipientEmail string, recipientAgentID uuid.UUID, template EmailTemplate, payload map[string]any) error {
	if s.repo == nil {
		return nil // No repo, skip
	}

	// Check cooldown to prevent spam
	if s.cooldownMinutes > 0 {
		recentlySent, err := s.repo.GetRecentEmailToRecipient(ctx, recipientAgentID, template, s.cooldownMinutes)
		if err != nil {
			log.Printf("[EMAIL] Error checking cooldown: %v", err)
		} else if recentlySent {
			log.Printf("[EMAIL] Skipping email to %s (cooldown active)", recipientEmail)
			return nil
		}
	}

	email := &QueuedEmail{
		ID:               uuid.New(),
		RecipientEmail:   recipientEmail,
		RecipientAgentID: recipientAgentID,
		Template:         template,
		Payload:          payload,
		Status:           EmailStatusPending,
		Attempts:         0,
		CreatedAt:        time.Now().UTC(),
	}

	return s.repo.QueueEmail(ctx, email)
}

// ProcessQueue processes pending emails from the queue.
func (s *Service) ProcessQueue(ctx context.Context) error {
	if s.client == nil || s.repo == nil {
		return nil
	}

	emails, err := s.repo.GetPendingEmails(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to get pending emails: %w", err)
	}

	for _, email := range emails {
		if err := s.sendEmail(ctx, email); err != nil {
			log.Printf("[EMAIL] Failed to send email %s: %v", email.ID, err)
			_ = s.repo.MarkEmailFailed(ctx, email.ID, err.Error())
		} else {
			log.Printf("[EMAIL] Sent email %s to %s", email.ID, email.RecipientEmail)
			_ = s.repo.MarkEmailSent(ctx, email.ID)
		}
	}

	return nil
}

// sendEmail sends a single email via SendGrid.
func (s *Service) sendEmail(ctx context.Context, email *QueuedEmail) error {
	data := s.buildEmailData(email)
	if data == nil {
		return fmt.Errorf("unknown template: %s", email.Template)
	}

	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail(data.ToName, email.RecipientEmail)
	message := mail.NewSingleEmail(from, data.Subject, to, data.TextContent, data.HTMLContent)

	response, err := s.client.Send(message)
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: %d - %s", response.StatusCode, response.Body)
	}

	return nil
}

// buildEmailData builds the email content from template and payload.
func (s *Service) buildEmailData(email *QueuedEmail) *EmailData {
	p := email.Payload
	getString := func(key string) string {
		if v, ok := p[key].(string); ok {
			return v
		}
		return ""
	}
	getFloat := func(key string) float64 {
		if v, ok := p[key].(float64); ok {
			return v
		}
		return 0
	}

	switch email.Template {
	case TemplateAgentRegistered:
		return &EmailData{
			ToName:  getString("agent_name"),
			Subject: "Welcome to SwarmMarket!",
			HTMLContent: fmt.Sprintf(`
				<h1>Welcome to SwarmMarket!</h1>
				<p>Hi %s,</p>
				<p>Your agent has been successfully registered on SwarmMarket - the autonomous agent marketplace.</p>
				<p><strong>Important:</strong> Save your API key securely. It will only be shown once!</p>
				<p><a href="%s/dashboard">Go to Dashboard</a></p>
				<p>Happy trading!</p>
			`, getString("agent_name"), s.baseURL),
			TextContent: fmt.Sprintf("Welcome to SwarmMarket, %s! Your agent has been registered.", getString("agent_name")),
		}

	case TemplateAgentClaimed:
		return &EmailData{
			ToName:  getString("owner_name"),
			Subject: "Agent Claimed Successfully",
			HTMLContent: fmt.Sprintf(`
				<h1>Agent Claimed!</h1>
				<p>Hi %s,</p>
				<p>You have successfully claimed the agent <strong>%s</strong>.</p>
				<p>You can now manage this agent from your dashboard.</p>
				<p><a href="%s/dashboard/agents">View Your Agents</a></p>
			`, getString("owner_name"), getString("agent_name"), s.baseURL),
			TextContent: fmt.Sprintf("You have claimed agent %s.", getString("agent_name")),
		}

	case TemplateNewMessage:
		return &EmailData{
			ToName:  getString("recipient_name"),
			Subject: fmt.Sprintf("New message from %s", getString("sender_name")),
			HTMLContent: fmt.Sprintf(`
				<h1>New Message</h1>
				<p>Hi %s,</p>
				<p>You have a new message from <strong>%s</strong>:</p>
				<blockquote style="border-left: 3px solid #22D3EE; padding-left: 12px; color: #666;">
					%s
				</blockquote>
				<p><a href="%s/dashboard/messages/%s">View Conversation</a></p>
			`, getString("recipient_name"), getString("sender_name"), getString("content_preview"), s.baseURL, getString("conversation_id")),
			TextContent: fmt.Sprintf("New message from %s: %s", getString("sender_name"), getString("content_preview")),
		}

	case TemplateOfferReceived:
		return &EmailData{
			ToName:  getString("requester_name"),
			Subject: fmt.Sprintf("New offer on your request: %s", getString("request_title")),
			HTMLContent: fmt.Sprintf(`
				<h1>New Offer Received!</h1>
				<p>Hi %s,</p>
				<p>You received a new offer on your request <strong>%s</strong>:</p>
				<ul>
					<li>From: %s</li>
					<li>Price: $%.2f %s</li>
				</ul>
				<p><a href="%s/marketplace/requests/%s">Review Offer</a></p>
			`, getString("requester_name"), getString("request_title"), getString("offerer_name"), getFloat("price_amount"), getString("price_currency"), s.baseURL, getString("request_id")),
			TextContent: fmt.Sprintf("New offer on %s from %s for $%.2f", getString("request_title"), getString("offerer_name"), getFloat("price_amount")),
		}

	case TemplateOfferAccepted:
		return &EmailData{
			ToName:  getString("offerer_name"),
			Subject: "Your offer was accepted!",
			HTMLContent: fmt.Sprintf(`
				<h1>Offer Accepted!</h1>
				<p>Hi %s,</p>
				<p>Great news! Your offer on <strong>%s</strong> has been accepted.</p>
				<p>Price: $%.2f %s</p>
				<p><a href="%s/dashboard/transactions">View Transaction</a></p>
			`, getString("offerer_name"), getString("request_title"), getFloat("price_amount"), getString("price_currency"), s.baseURL),
			TextContent: fmt.Sprintf("Your offer on %s was accepted!", getString("request_title")),
		}

	case TemplateNewBid:
		return &EmailData{
			ToName:  getString("seller_name"),
			Subject: fmt.Sprintf("New bid on your auction: %s", getString("auction_title")),
			HTMLContent: fmt.Sprintf(`
				<h1>New Bid!</h1>
				<p>Hi %s,</p>
				<p>Your auction <strong>%s</strong> received a new bid:</p>
				<ul>
					<li>Bidder: %s</li>
					<li>Amount: $%.2f %s</li>
				</ul>
				<p><a href="%s/marketplace/auctions/%s">View Auction</a></p>
			`, getString("seller_name"), getString("auction_title"), getString("bidder_name"), getFloat("bid_amount"), getString("currency"), s.baseURL, getString("auction_id")),
			TextContent: fmt.Sprintf("New bid of $%.2f on %s", getFloat("bid_amount"), getString("auction_title")),
		}

	case TemplateOutbid:
		return &EmailData{
			ToName:  getString("bidder_name"),
			Subject: fmt.Sprintf("You've been outbid on: %s", getString("auction_title")),
			HTMLContent: fmt.Sprintf(`
				<h1>You've Been Outbid!</h1>
				<p>Hi %s,</p>
				<p>Someone has outbid you on <strong>%s</strong>.</p>
				<ul>
					<li>Your bid: $%.2f</li>
					<li>New highest bid: $%.2f</li>
				</ul>
				<p><a href="%s/marketplace/auctions/%s">Place a Higher Bid</a></p>
			`, getString("bidder_name"), getString("auction_title"), getFloat("your_bid"), getFloat("new_highest_bid"), s.baseURL, getString("auction_id")),
			TextContent: fmt.Sprintf("You've been outbid on %s. New highest: $%.2f", getString("auction_title"), getFloat("new_highest_bid")),
		}

	case TemplateAuctionWon:
		return &EmailData{
			ToName:  getString("winner_name"),
			Subject: fmt.Sprintf("Congratulations! You won: %s", getString("auction_title")),
			HTMLContent: fmt.Sprintf(`
				<h1>Congratulations!</h1>
				<p>Hi %s,</p>
				<p>You won the auction for <strong>%s</strong>!</p>
				<p>Winning bid: $%.2f %s</p>
				<p><a href="%s/dashboard/transactions">Complete Transaction</a></p>
			`, getString("winner_name"), getString("auction_title"), getFloat("winning_bid"), getString("currency"), s.baseURL),
			TextContent: fmt.Sprintf("You won %s for $%.2f!", getString("auction_title"), getFloat("winning_bid")),
		}

	case TemplateAuctionEnded:
		return &EmailData{
			ToName:  getString("seller_name"),
			Subject: fmt.Sprintf("Your auction has ended: %s", getString("auction_title")),
			HTMLContent: fmt.Sprintf(`
				<h1>Auction Ended</h1>
				<p>Hi %s,</p>
				<p>Your auction <strong>%s</strong> has ended.</p>
				<p>Final price: $%.2f %s</p>
				<p>Winner: %s</p>
				<p><a href="%s/dashboard/transactions">View Transaction</a></p>
			`, getString("seller_name"), getString("auction_title"), getFloat("final_price"), getString("currency"), getString("winner_name"), s.baseURL),
			TextContent: fmt.Sprintf("Your auction %s ended. Final price: $%.2f", getString("auction_title"), getFloat("final_price")),
		}

	case TemplateListingPurchased:
		return &EmailData{
			ToName:  getString("seller_name"),
			Subject: fmt.Sprintf("Your listing was purchased: %s", getString("listing_title")),
			HTMLContent: fmt.Sprintf(`
				<h1>Listing Purchased!</h1>
				<p>Hi %s,</p>
				<p>Your listing <strong>%s</strong> has been purchased!</p>
				<ul>
					<li>Buyer: %s</li>
					<li>Price: $%.2f %s</li>
				</ul>
				<p><a href="%s/dashboard/transactions">View Transaction</a></p>
			`, getString("seller_name"), getString("listing_title"), getString("buyer_name"), getFloat("price"), getString("currency"), s.baseURL),
			TextContent: fmt.Sprintf("%s purchased %s for $%.2f", getString("buyer_name"), getString("listing_title"), getFloat("price")),
		}

	case TemplateTransactionComplete:
		return &EmailData{
			ToName:  getString("recipient_name"),
			Subject: "Transaction completed!",
			HTMLContent: fmt.Sprintf(`
				<h1>Transaction Complete!</h1>
				<p>Hi %s,</p>
				<p>Your transaction has been completed successfully.</p>
				<p>Amount: $%.2f %s</p>
				<p><a href="%s/dashboard/transactions/%s">View Details</a></p>
			`, getString("recipient_name"), getFloat("amount"), getString("currency"), s.baseURL, getString("transaction_id")),
			TextContent: fmt.Sprintf("Transaction completed for $%.2f", getFloat("amount")),
		}

	case TemplateDisputeOpened:
		return &EmailData{
			ToName:  getString("recipient_name"),
			Subject: "A dispute has been opened",
			HTMLContent: fmt.Sprintf(`
				<h1>Dispute Opened</h1>
				<p>Hi %s,</p>
				<p>A dispute has been opened on one of your transactions.</p>
				<p>Reason: %s</p>
				<p><a href="%s/dashboard/transactions/%s">View Transaction</a></p>
			`, getString("recipient_name"), getString("reason"), s.baseURL, getString("transaction_id")),
			TextContent: fmt.Sprintf("Dispute opened: %s", getString("reason")),
		}

	default:
		return nil
	}
}

// IsConfigured returns true if the email service is properly configured.
func (s *Service) IsConfigured() bool {
	return s.client != nil
}
