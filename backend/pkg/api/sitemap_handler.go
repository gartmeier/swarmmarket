package api

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/auction"
	"github.com/digi604/swarmmarket/backend/internal/marketplace"
)

// SitemapHandler handles sitemap generation.
type SitemapHandler struct {
	marketplaceService *marketplace.Service
	auctionService     *auction.Service
	baseURL            string
}

// NewSitemapHandler creates a new sitemap handler.
func NewSitemapHandler(marketplaceService *marketplace.Service, auctionService *auction.Service, baseURL string) *SitemapHandler {
	return &SitemapHandler{
		marketplaceService: marketplaceService,
		auctionService:     auctionService,
		baseURL:            baseURL,
	}
}

// URLSet represents the root element of a sitemap.
type URLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// SitemapURL represents a single URL entry in the sitemap.
type SitemapURL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// GenerateSitemap generates the sitemap XML.
// GET /sitemap.xml
func (h *SitemapHandler) GenerateSitemap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	urls := []SitemapURL{
		// Static pages
		{Loc: h.baseURL, ChangeFreq: "daily", Priority: 1.0},
		{Loc: h.baseURL + "/marketplace", ChangeFreq: "hourly", Priority: 0.9},
	}

	// Add listings
	if h.marketplaceService != nil {
		activeStatus := marketplace.ListingStatusActive
		listings, err := h.marketplaceService.SearchListings(ctx, marketplace.SearchListingsParams{
			Status: &activeStatus,
			Limit:  1000,
		})
		if err == nil {
			for _, listing := range listings.Items {
				slug := listing.Slug
				if slug == "" {
					slug = listing.ID.String()
				}
				urls = append(urls, SitemapURL{
					Loc:        fmt.Sprintf("%s/marketplace/listings/%s", h.baseURL, slug),
					LastMod:    listing.UpdatedAt.Format(time.RFC3339),
					ChangeFreq: "daily",
					Priority:   0.8,
				})
			}
		}

		// Add requests
		openStatus := marketplace.RequestStatusOpen
		requests, err := h.marketplaceService.SearchRequests(ctx, marketplace.SearchRequestsParams{
			Status: &openStatus,
			Limit:  1000,
		})
		if err == nil {
			for _, request := range requests.Items {
				slug := request.Slug
				if slug == "" {
					slug = request.ID.String()
				}
				urls = append(urls, SitemapURL{
					Loc:        fmt.Sprintf("%s/marketplace/requests/%s", h.baseURL, slug),
					LastMod:    request.UpdatedAt.Format(time.RFC3339),
					ChangeFreq: "daily",
					Priority:   0.8,
				})
			}
		}
	}

	// Add auctions
	if h.auctionService != nil {
		auctionActiveStatus := auction.AuctionStatusActive
		auctions, err := h.auctionService.SearchAuctions(ctx, auction.SearchAuctionsParams{
			Status: &auctionActiveStatus,
			Limit:  1000,
		})
		if err == nil {
			for _, auc := range auctions.Auctions {
				slug := auc.Slug
				if slug == "" {
					slug = auc.ID.String()
				}
				urls = append(urls, SitemapURL{
					Loc:        fmt.Sprintf("%s/marketplace/auctions/%s", h.baseURL, slug),
					LastMod:    auc.UpdatedAt.Format(time.RFC3339),
					ChangeFreq: "hourly",
					Priority:   0.8,
				})
			}
		}
	}

	urlSet := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	encoder.Encode(urlSet)
}

// GenerateRobotsTxt generates robots.txt with sitemap reference.
// GET /robots.txt
func (h *SitemapHandler) GenerateRobotsTxt(w http.ResponseWriter, r *http.Request) {
	robotsTxt := fmt.Sprintf(`User-agent: *
Allow: /
Allow: /marketplace
Allow: /marketplace/listings/
Allow: /marketplace/requests/
Allow: /marketplace/auctions/

Disallow: /dashboard/
Disallow: /api/

Sitemap: %s/sitemap.xml
`, h.baseURL)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(robotsTxt))
}
