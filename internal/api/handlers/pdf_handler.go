package handlers

import (
	"bytes"
	"expense-tracker/internal/api/middleware"
	"expense-tracker/internal/models"
	"expense-tracker/internal/services"
	"expense-tracker/pkg/response"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v2"
)

// PDFHandler generates a monthly financial report PDF.
type PDFHandler struct {
	txSvc services.TransactionService
}

func NewPDFHandler(txSvc services.TransactionService) *PDFHandler {
	return &PDFHandler{txSvc: txSvc}
}

// colour helpers — fpdf uses 0-255 RGB
type rgb struct{ r, g, b uint8 }

var (
	colIncome  = rgb{34, 197, 94}   // green-500
	colExpense = rgb{239, 68, 68}   // red-500
	colPrimary = rgb{99, 102, 241}  // indigo-500
	colGray    = rgb{107, 114, 128} // gray-500
	colLight   = rgb{249, 250, 251} // gray-50
	colMid     = rgb{243, 244, 246} // gray-100
	colBorder  = rgb{229, 231, 235} // gray-200
	colWhite   = rgb{255, 255, 255}
	colBlack   = rgb{17, 24, 39}    // gray-900
)

func setFill(pdf *fpdf.Fpdf, c rgb) { pdf.SetFillColor(int(c.r), int(c.g), int(c.b)) }
func setDraw(pdf *fpdf.Fpdf, c rgb) { pdf.SetDrawColor(int(c.r), int(c.g), int(c.b)) }
func setText(pdf *fpdf.Fpdf, c rgb) { pdf.SetTextColor(int(c.r), int(c.g), int(c.b)) }

func safeStr(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 128 {
			b.WriteRune(r)
		} else {
			b.WriteRune('?')
		}
	}
	return b.String()
}

func fmtAmount(amount float64, currency string) string {
	return fmt.Sprintf("%s %.2f", currency, amount)
}

func (h *PDFHandler) GeneratePDF(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	// Parse date range
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	currency := c.Query("currency", "USD")

	// Fetch data
	summary, err := h.txSvc.GetSummary(c.Context(), userID, from, to)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch summary")
	}

	breakdown, err := h.txSvc.GetCategoryBreakdown(c.Context(), userID, from, to, models.TransactionExpense)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch breakdown")
	}

	filter := &models.TransactionFilter{Page: 1, PerPage: 500}
	filter.FromDate = &from
	filter.ToDate = &to
	transactions, _, err := h.txSvc.List(c.Context(), userID, filter)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch transactions")
	}

	// ── Build PDF ───────────────────────────────────────────────────────────
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pageW, _ := pdf.GetPageSize()
	usableW := pageW - 30 // left + right margins

	period := fmt.Sprintf("%s %d", from.Month().String(), from.Year())
	monthRange := fmt.Sprintf("%s – %s", from.Format("Jan 2"), to.Format("Jan 2, 2006"))

	// ── Header ──────────────────────────────────────────────────────────────
	setText(pdf, colBlack)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.Cell(100, 9, "ExpenseTracker")

	pdf.SetFont("Helvetica", "B", 13)
	setText(pdf, colPrimary)
	pdf.SetXY(pageW-60, pdf.GetY())
	pdf.CellFormat(45, 9, period, "", 0, "R", false, 0, "")

	pdf.Ln(9)
	pdf.SetFont("Helvetica", "", 9)
	setText(pdf, colGray)
	pdf.Cell(100, 5, "Monthly Financial Report")
	pdf.SetXY(pageW-60, pdf.GetY())
	pdf.CellFormat(45, 5, monthRange, "", 0, "R", false, 0, "")

	pdf.Ln(7)
	setDraw(pdf, colBorder)
	pdf.SetLineWidth(0.3)
	pdf.Line(15, pdf.GetY(), pageW-15, pdf.GetY())
	pdf.Ln(6)

	// ── Summary cards ────────────────────────────────────────────────────────
	setText(pdf, colGray)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.Cell(usableW, 5, "FINANCIAL SUMMARY")
	pdf.Ln(7)

	cardW := (usableW - 6) / 3
	cardH := 20.0
	cards := []struct {
		label  string
		amount float64
		col    rgb
		bg     rgb
	}{
		{"Total Income", summary.TotalIncome, colIncome, rgb{240, 253, 244}},
		{"Total Expenses", summary.TotalExpense, colExpense, rgb{254, 242, 242}},
		{"Net Balance", summary.NetBalance, colPrimary, rgb{238, 242, 255}},
	}
	if summary.NetBalance < 0 {
		cards[2].col = colExpense
		cards[2].bg = rgb{254, 242, 242}
	}

	startX := 15.0
	for i, card := range cards {
		x := startX + float64(i)*(cardW+3)
		y := pdf.GetY()
		setFill(pdf, card.bg)
		setDraw(pdf, colBorder)
		pdf.RoundedRect(x, y, cardW, cardH, 2, "1234", "FD")

		pdf.SetFont("Helvetica", "", 8)
		setText(pdf, colGray)
		pdf.SetXY(x+3, y+3)
		pdf.CellFormat(cardW-6, 5, card.label, "", 2, "L", false, 0, "")

		pdf.SetFont("Helvetica", "B", 11)
		setText(pdf, card.col)
		pdf.SetXY(x+3, y+9)
		sign := ""
		if card.label == "Net Balance" && card.amount > 0 {
			sign = "+"
		} else if card.label == "Net Balance" && card.amount < 0 {
			sign = "-"
			card.amount = math.Abs(card.amount)
		}
		pdf.CellFormat(cardW-6, 7, sign+fmtAmount(card.amount, currency), "", 0, "L", false, 0, "")
	}
	pdf.SetY(pdf.GetY() + cardH + 8)

	// ── Category breakdown ───────────────────────────────────────────────────
	if len(breakdown) > 0 {
		setText(pdf, colGray)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.Cell(usableW, 5, "SPENDING BY CATEGORY")
		pdf.Ln(7)

		maxShow := 8
		if len(breakdown) < maxShow {
			maxShow = len(breakdown)
		}

		barFullW := usableW * 0.55
		for _, item := range breakdown[:maxShow] {
			y := pdf.GetY()

			// Category name
			pdf.SetFont("Helvetica", "", 9)
			setText(pdf, colBlack)
			pdf.SetXY(15, y+1)
			name := safeStr(item.CategoryName)
			if len(name) > 22 {
				name = name[:22] + "…"
			}
			pdf.CellFormat(48, 5, name, "", 0, "L", false, 0, "")

			// Progress bar background
			barX := 65.0
			barY := y + 2.5
			barH := 3.5
			setFill(pdf, colMid)
			setDraw(pdf, colMid)
			pdf.RoundedRect(barX, barY, barFullW, barH, 1.5, "1234", "FD")

			// Progress bar fill
			fillW := barFullW * (item.Percentage / 100)
			if fillW < 1 {
				fillW = 1
			}
			// parse hex color from item
			fillCol := colPrimary
			if len(item.CategoryColor) == 7 {
				r, _ := strconv.ParseUint(item.CategoryColor[1:3], 16, 8)
				g, _ := strconv.ParseUint(item.CategoryColor[3:5], 16, 8)
				b, _ := strconv.ParseUint(item.CategoryColor[5:7], 16, 8)
				fillCol = rgb{uint8(r), uint8(g), uint8(b)}
			}
			setFill(pdf, fillCol)
			setDraw(pdf, fillCol)
			pdf.RoundedRect(barX, barY, fillW, barH, 1.5, "1234", "FD")

			// Percentage
			pdf.SetFont("Helvetica", "", 8)
			setText(pdf, colGray)
			pdf.SetXY(barX+barFullW+3, y)
			pdf.CellFormat(14, 7, fmt.Sprintf("%.0f%%", item.Percentage), "", 0, "R", false, 0, "")

			// Amount
			pdf.SetFont("Helvetica", "B", 8)
			setText(pdf, colBlack)
			pdf.SetXY(barX+barFullW+17, y)
			pdf.CellFormat(usableW-(barX-15+barFullW+17), 7, fmtAmount(item.Total, currency), "", 0, "R", false, 0, "")

			pdf.Ln(8)
		}
		pdf.Ln(3)
	}

	// ── Transactions table ────────────────────────────────────────────────────
	setText(pdf, colGray)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.Cell(usableW, 5, "TRANSACTIONS")
	pdf.Ln(7)

	// Table header
	colWidths := []float64{22, 78, 40, 40}
	headers := []string{"Date", "Description", "Category", "Amount"}
	aligns := []string{"L", "L", "L", "R"}

	setFill(pdf, colMid)
	setDraw(pdf, colBorder)
	pdf.SetFont("Helvetica", "B", 8)
	setText(pdf, colGray)
	for i, hdr := range headers {
		pdf.CellFormat(colWidths[i], 7, hdr, "B", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(7)

	// Rows
	pdf.SetFont("Helvetica", "", 8)
	for idx, tx := range transactions {
		if pdf.GetY() > 265 {
			pdf.AddPage()
		}

		isIncome := tx.Type == models.TransactionIncome
		if idx%2 == 0 {
			setFill(pdf, colLight)
		} else {
			setFill(pdf, colWhite)
		}

		dateStr := tx.Date.Format("Jan 02")
		desc := safeStr(tx.Description)
		if len(desc) > 38 {
			desc = desc[:38] + "…"
		}
		catName := "—"
		if tx.Category.Name != "" {
			catName = safeStr(tx.Category.Name)
			if len(catName) > 18 {
				catName = catName[:18] + "…"
			}
		}

		sign := "-"
		amtCol := colExpense
		if isIncome {
			sign = "+"
			amtCol = colIncome
		}
		amtStr := sign + fmtAmount(tx.Amount, currency)

		setText(pdf, colGray)
		pdf.CellFormat(colWidths[0], 6.5, dateStr, "", 0, "L", true, 0, "")
		setText(pdf, colBlack)
		pdf.CellFormat(colWidths[1], 6.5, desc, "", 0, "L", true, 0, "")
		setText(pdf, colGray)
		pdf.CellFormat(colWidths[2], 6.5, catName, "", 0, "L", true, 0, "")
		setText(pdf, amtCol)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.CellFormat(colWidths[3], 6.5, amtStr, "", 0, "R", true, 0, "")
		pdf.SetFont("Helvetica", "", 8)
		pdf.Ln(6.5)
	}

	if len(transactions) == 0 {
		setText(pdf, colGray)
		pdf.CellFormat(usableW, 10, "No transactions for this period.", "", 0, "C", false, 0, "")
		pdf.Ln(10)
	}

	// ── Footer ────────────────────────────────────────────────────────────────
	pdf.Ln(5)
	setDraw(pdf, colBorder)
	pdf.Line(15, pdf.GetY(), pageW-15, pdf.GetY())
	pdf.Ln(4)
	setText(pdf, colGray)
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(usableW, 5, "Generated by ExpenseTracker · "+time.Now().Format("January 2, 2006 15:04"), "", 0, "C", false, 0, "")

	// ── Output ────────────────────────────────────────────────────────────────
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return response.InternalServerError(c, "failed to generate PDF")
	}

	filename := fmt.Sprintf("report-%s-%d.pdf", strings.ToLower(from.Month().String()), from.Year())
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.Send(buf.Bytes())
}
