package service

import (
	"fmt"
	"image/color"
	"strings"

	"hrbackend/internal/domain"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// Color constants for professional branding
var (
	colorPrimary   = color.RGBA{R: 25, G: 63, B: 141, A: 255}   // Professional blue
	colorSecondary = color.RGBA{R: 52, G: 152, B: 219, A: 255}  // Light blue
	colorAccent    = color.RGBA{R: 44, G: 62, B: 80, A: 255}    // Dark gray
	colorBorder    = color.RGBA{R: 189, G: 195, B: 199, A: 255} // Light gray
	colorSuccess   = color.RGBA{R: 27, G: 148, B: 60, A: 255}   // Green
	colorText      = color.RGBA{R: 44, G: 62, B: 80, A: 255}    // Dark text
	colorTextLight = color.RGBA{R: 127, G: 140, B: 141, A: 255} // Light text
)

func buildPayrollMonthDetailPDF(detail *domain.PayrollMonthDetail) ([]byte, error) {
	if detail == nil {
		return nil, fmt.Errorf("payroll detail is required")
	}

	cfg := config.NewBuilder().
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(15).
		Build()

	m := maroto.New(cfg)

	// Header section
	addCompanyHeader(m)
	addCompanyDivider(m)

	// Title and metadata
	addPayrollTitle(m, detail)
	m.AddRow(5, col.New(12).Add(text.New("", props.Text{}))) // Spacing

	// Employee information box
	addEmployeeInfoSection(m, detail)
	m.AddRow(3, col.New(12).Add(text.New("", props.Text{}))) // Spacing

	// Period details based on type
	if detail.PayPeriod != nil {
		addPayPeriodDetails(m, detail.PayPeriod)
		addPayPeriodLineItems(m, detail.PayPeriod)
	} else if detail.Preview != nil {
		addPayrollPreviewDetails(m, detail.Preview)
		addPayrollPreviewLineItems(m, detail.Preview)
	}

	m.AddRow(5, col.New(12).Add(text.New("", props.Text{}))) // Spacing
	addFooter(m)

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate payroll pdf: %w", err)
	}

	return document.GetBytes(), nil
}

func addCompanyHeader(m core.Maroto) {
	// Header background with color
	m.AddRow(20, col.New(12).Add(
		text.New("COMPANY PAYROLL", props.Text{
			Style: fontstyle.Bold,
			Size:  24,
			Color: rgbaToPropsColor(colorPrimary),
			Align: align.Left,
		}),
	))

	// Company info line
	m.AddRow(4, col.New(12).Add(
		text.New("Professional Payroll Statement", props.Text{
			Size:  10,
			Color: rgbaToPropsColor(colorTextLight),
			Align: align.Left,
		}),
	))
}

func addCompanyDivider(m core.Maroto) {
	// Add a visual divider
	m.AddRow(2, col.New(12).Add(
		text.New("", props.Text{
			Size: 1,
		}),
	))
}

func addPayrollTitle(m core.Maroto, detail *domain.PayrollMonthDetail) {
	// Document type
	m.AddRow(6, col.New(6).Add(
		text.New("SALARY OVERVIEW", props.Text{
			Style: fontstyle.Bold,
			Size:  14,
			Color: rgbaToPropsColor(colorAccent),
		}),
	), col.New(6).Add(
		text.New(fmt.Sprintf("Period: %s", detail.Month.Format("January 2006")), props.Text{
			Size:  11,
			Color: rgbaToPropsColor(colorText),
			Align: align.Right,
		}),
	))
}

func addEmployeeInfoSection(m core.Maroto, detail *domain.PayrollMonthDetail) {
	// Employee info in a structured format
	m.AddRow(5, col.New(6).Add(
		text.New("Employee Name", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	), col.New(6).Add(
		text.New("Data Source", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	))

	m.AddRow(5, col.New(6).Add(
		text.New(detail.EmployeeName, props.Text{
			Size:  11,
			Color: rgbaToPropsColor(colorText),
		}),
	), col.New(6).Add(
		text.New(strings.ToUpper(detail.DataSource), props.Text{
			Size:  11,
			Color: rgbaToPropsColor(colorText),
		}),
	))
}

func addPayPeriodDetails(m core.Maroto, period *domain.PayPeriod) {
	// Section title
	m.AddRow(6, col.New(12).Add(
		text.New("PERIOD INFORMATION", props.Text{
			Style: fontstyle.Bold,
			Size:  11,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	))

	// Period status badge and dates
	m.AddRow(5, col.New(3).Add(
		text.New("Status", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	), col.New(3).Add(
		text.New("Period Start", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	), col.New(3).Add(
		text.New("Period End", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	), col.New(3).Add(
		text.New("Created", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	))

	statusColor := getStatusColor(period.Status)
	m.AddRow(5, col.New(3).Add(
		text.New(strings.ToUpper(period.Status), props.Text{
			Size:  10,
			Color: rgbaToPropsColor(statusColor),
			Style: fontstyle.Bold,
		}),
	), col.New(3).Add(
		text.New(period.PeriodStart.Format("2006-01-02"), props.Text{
			Size:  10,
			Color: rgbaToPropsColor(colorText),
		}),
	), col.New(3).Add(
		text.New(period.PeriodEnd.Format("2006-01-02"), props.Text{
			Size:  10,
			Color: rgbaToPropsColor(colorText),
		}),
	), col.New(3).Add(
		text.New(period.CreatedAt.Format("2006-01-02"), props.Text{
			Size:  10,
			Color: rgbaToPropsColor(colorText),
		}),
	))

	m.AddRow(3, col.New(12).Add(text.New("", props.Text{}))) // Spacing

	// Earnings summary table
	addEarningsSummaryTable(m, period.BaseGrossAmount, period.IrregularGrossAmount, period.GrossAmount)
	m.AddRow(3, col.New(12).Add(text.New("", props.Text{}))) // Spacing
}

func addPayrollPreviewDetails(m core.Maroto, preview *domain.PayrollPreview) {
	// Section title
	m.AddRow(6, col.New(12).Add(
		text.New("PAYROLL PREVIEW", props.Text{
			Style: fontstyle.Bold,
			Size:  11,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	))

	// Preview dates and hours
	m.AddRow(5, col.New(4).Add(
		text.New("Period Start", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	), col.New(4).Add(
		text.New("Period End", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	), col.New(4).Add(
		text.New("Total Worked", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	))

	m.AddRow(5, col.New(4).Add(
		text.New(preview.PeriodStart.Format("2006-01-02"), props.Text{
			Size:  10,
			Color: rgbaToPropsColor(colorText),
		}),
	), col.New(4).Add(
		text.New(preview.PeriodEnd.Format("2006-01-02"), props.Text{
			Size:  10,
			Color: rgbaToPropsColor(colorText),
		}),
	), col.New(4).Add(
		text.New(fmt.Sprintf("%d minutes", preview.TotalWorkedMinutes), props.Text{
			Size:  10,
			Color: rgbaToPropsColor(colorText),
		}),
	))

	m.AddRow(3, col.New(12).Add(text.New("", props.Text{}))) // Spacing

	// Earnings summary table
	addEarningsSummaryTable(m, preview.BaseGrossAmount, preview.IrregularGrossAmount, preview.GrossAmount)
	m.AddRow(3, col.New(12).Add(text.New("", props.Text{}))) // Spacing
}

func addEarningsSummaryTable(m core.Maroto, baseAmount, irregularAmount, totalAmount float64) {
	// Summary box with key earnings information
	m.AddRow(6, col.New(12).Add(
		text.New("EARNINGS SUMMARY", props.Text{
			Style: fontstyle.Bold,
			Size:  11,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	))

	m.AddRow(5, col.New(6).Add(
		text.New("Base Salary", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	), col.New(6).Add(
		text.New("Irregular Hours (ORT)", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorTextLight),
		}),
	))

	m.AddRow(5, col.New(6).Add(
		text.New(fmt.Sprintf("€ %.2f", baseAmount), props.Text{
			Size:  11,
			Color: rgbaToPropsColor(colorText),
		}),
	), col.New(6).Add(
		text.New(fmt.Sprintf("€ %.2f", irregularAmount), props.Text{
			Size:  11,
			Color: rgbaToPropsColor(colorText),
		}),
	))

	m.AddRow(3, col.New(12).Add(text.New("", props.Text{}))) // Spacing

	// Total amount in highlighted box
	m.AddRow(7, col.New(12).Add(
		text.New(fmt.Sprintf("GROSS TOTAL: € %.2f", totalAmount), props.Text{
			Style: fontstyle.Bold,
			Size:  13,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	))
}

func addPayPeriodLineItems(m core.Maroto, period *domain.PayPeriod) {
	if len(period.LineItems) == 0 {
		m.AddRow(5, col.New(12).Add(
			text.New("No line items", props.Text{
				Size:  10,
				Color: rgbaToPropsColor(colorTextLight),
				Style: fontstyle.Italic,
			}),
		))
		return
	}

	// Section title
	m.AddRow(6, col.New(12).Add(
		text.New("DETAILED LINE ITEMS", props.Text{
			Style: fontstyle.Bold,
			Size:  11,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	))

	// Display line items with proper formatting
	m.AddRow(4, col.New(2).Add(
		text.New("Date", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	), col.New(2).Add(
		text.New("Type", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	), col.New(2).Add(
		text.New("Minutes", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	), col.New(3).Add(
		text.New("Base Amount", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	), col.New(3).Add(
		text.New("Premium Amount", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	))

	itemCount := 0
	for _, line := range period.LineItems {
		if itemCount >= 30 {
			m.AddRow(4, col.New(12).Add(
				text.New("... (output truncated - showing first 30 items)", props.Text{
					Size:  9,
					Color: rgbaToPropsColor(colorTextLight),
					Style: fontstyle.Italic,
				}),
			))
			break
		}
		m.AddRow(4, col.New(2).Add(
			text.New(line.WorkDate.Format("01-02"), props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		), col.New(2).Add(
			text.New(line.LineType, props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		), col.New(2).Add(
			text.New(fmt.Sprintf("%.0f", line.MinutesWorked), props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		), col.New(3).Add(
			text.New(fmt.Sprintf("€ %.2f", line.BaseAmount), props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		), col.New(3).Add(
			text.New(fmt.Sprintf("€ %.2f", line.PremiumAmount), props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		))
		itemCount++
	}
}

func addPayrollPreviewLineItems(m core.Maroto, preview *domain.PayrollPreview) {
	if len(preview.LineItems) == 0 {
		m.AddRow(5, col.New(12).Add(
			text.New("No preview items", props.Text{
				Size:  10,
				Color: rgbaToPropsColor(colorTextLight),
				Style: fontstyle.Italic,
			}),
		))
		return
	}

	// Section title
	m.AddRow(6, col.New(12).Add(
		text.New("PREVIEW LINE ITEMS", props.Text{
			Style: fontstyle.Bold,
			Size:  11,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	))

	// Display line items with proper formatting
	m.AddRow(4, col.New(2).Add(
		text.New("Date", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	), col.New(2).Add(
		text.New("Hour Type", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	), col.New(2).Add(
		text.New("Minutes", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	), col.New(3).Add(
		text.New("Base Amount", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	), col.New(3).Add(
		text.New("Premium Amount", props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Color: rgbaToPropsColor(colorPrimary),
		}),
	))

	itemCount := 0
	for _, line := range preview.LineItems {
		if itemCount >= 30 {
			m.AddRow(4, col.New(12).Add(
				text.New("... (output truncated - showing first 30 items)", props.Text{
					Size:  9,
					Color: rgbaToPropsColor(colorTextLight),
					Style: fontstyle.Italic,
				}),
			))
			break
		}
		m.AddRow(4, col.New(2).Add(
			text.New(line.WorkDate.Format("01-02"), props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		), col.New(2).Add(
			text.New(line.HourType, props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		), col.New(2).Add(
			text.New(fmt.Sprintf("%d", line.MinutesWorked), props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		), col.New(3).Add(
			text.New(fmt.Sprintf("€ %.2f", line.BaseAmount), props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		), col.New(3).Add(
			text.New(fmt.Sprintf("€ %.2f", line.PremiumAmount), props.Text{
				Size:  9,
				Color: rgbaToPropsColor(colorText),
			}),
		))
		itemCount++
	}
}

func addFooter(m core.Maroto) {
	// Footer section
	m.AddRow(1, col.New(12).Add(text.New("", props.Text{}))) // Spacing

	m.AddRow(4, col.New(12).Add(
		text.New("This is a confidential payroll document. Please handle with care.", props.Text{
			Size:  8,
			Color: rgbaToPropsColor(colorTextLight),
			Style: fontstyle.Italic,
			Align: align.Center,
		}),
	))

	m.AddRow(3, col.New(12).Add(
		text.New("Generated by HR Management System", props.Text{
			Size:  8,
			Color: rgbaToPropsColor(colorTextLight),
			Align: align.Center,
		}),
	))
}

func getStatusColor(status string) color.RGBA {
	switch strings.ToLower(status) {
	case "paid":
		return colorSuccess
	case "draft":
		return colorSecondary
	default:
		return colorAccent
	}
}

// Helper function to convert color.RGBA to *props.Color
func rgbaToPropsColor(c color.RGBA) *props.Color {
	return &props.Color{
		Red:   int(c.R),
		Green: int(c.G),
		Blue:  int(c.B),
	}
}
