package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/protob/event-sensor/ticketmaster"
	"github.com/spf13/cobra"
)

var apiKey string
var outputJSON bool
var filterParking bool
var filterDuplicates bool

func main() {
	var rootCmd = &cobra.Command{
		Use:   "tm-cli",
		Short: "Ticketmaster API CLI for debugging",
	}

	rootCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", "Ticketmaster API key (or set TICKETMASTER_API_KEY env)")

	var searchCmd = &cobra.Command{
		Use:   "search [keyword]",
		Short: "Search events by keyword",
		Args:  cobra.ExactArgs(1),
		Run:   runSearch,
	}
	searchCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output raw JSON")
	searchCmd.Flags().BoolVarP(&filterParking, "no-parking", "p", false, "Filter out parking/add-on events")
	searchCmd.Flags().BoolVarP(&filterDuplicates, "no-dupes", "d", false, "Filter duplicate event names")

	var analyzeCmd = &cobra.Command{
		Use:   "analyze [keyword]",
		Short: "Analyze events and show statistics",
		Args:  cobra.ExactArgs(1),
		Run:   runAnalyze,
	}

	rootCmd.AddCommand(searchCmd, analyzeCmd)

	if apiKey == "" {
		apiKey = os.Getenv("TICKETMASTER_API_KEY")
	}
	if apiKey == "" {
		apiKey = readAgenixKey()
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func readAgenixKey() string {
	data, err := os.ReadFile("/run/agenix/ticketmaster-api-key")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func getAPIKey() string {
	if apiKey != "" {
		return apiKey
	}
	if key := os.Getenv("TICKETMASTER_API_KEY"); key != "" {
		return key
	}
	if key := readAgenixKey(); key != "" {
		return key
	}
	log.Fatal("No API key provided. Use --api-key, TICKETMASTER_API_KEY env, or /run/agenix/ticketmaster-api-key")
	return ""
}

func runSearch(cmd *cobra.Command, args []string) {
	keyword := args[0]
	client := ticketmaster.NewClient(getAPIKey())

	resp, err := client.SearchEvents(keyword)
	if err != nil {
		log.Fatalf("API error: %v", err)
	}

	if outputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(resp)
		return
	}

	if resp.Embedded == nil {
		fmt.Println("No events found")
		return
	}

	events := resp.Embedded.Events
	original := len(events)

	if filterParking {
		events = filterParkingEvents(events)
	}
	if filterDuplicates {
		events = filterDuplicateEvents(events)
	}

	fmt.Printf("Found %d events", len(events))
	if filterParking || filterDuplicates {
		fmt.Printf(" (filtered from %d)", original)
	}
	fmt.Print("\n\n")

	for i, e := range events {
		venue := "unknown"
		country := "??"
		if e.Embedded != nil && len(e.Embedded.Venues) > 0 {
			v := e.Embedded.Venues[0]
			venue = v.Name
			country = v.Country.CountryCode
		}
		date := "unknown"
		if e.Dates.Start.LocalDate != "" {
			date = e.Dates.Start.LocalDate
		}
		fmt.Printf("%3d. [%s] %s - %s (%s)\n", i+1, date, e.Name, venue, country)
		fmt.Printf("      ID: %s | Type: %s\n", e.ID, e.Type)
		if len(e.Classifications) > 0 {
			fmt.Printf("      Segment: %s | Genre: %s\n",
				e.Classifications[0].Segment.Name,
				e.Classifications[0].Genre.Name)
		}
		fmt.Println()
	}
}

func runAnalyze(cmd *cobra.Command, args []string) {
	keyword := args[0]
	client := ticketmaster.NewClient(getAPIKey())

	resp, err := client.SearchEvents(keyword)
	if err != nil {
		log.Fatalf("API error: %v", err)
	}

	if resp.Embedded == nil {
		fmt.Println("No events found")
		return
	}

	events := resp.Embedded.Events

	fmt.Printf("=== Analysis for '%s' ===\n\n", keyword)
	fmt.Printf("Total events: %d\n\n", len(events))

	byType := make(map[string]int)
	byCountry := make(map[string]int)
	bySegment := make(map[string]int)
	byGenre := make(map[string]int)
	nameCounts := make(map[string]int)
	parkingCount := 0
	addonCount := 0

	for _, e := range events {
		byType[e.Type]++

		if e.Embedded != nil && len(e.Embedded.Venues) > 0 {
			cc := e.Embedded.Venues[0].Country.CountryCode
			byCountry[cc]++
		}

		if len(e.Classifications) > 0 {
			bySegment[e.Classifications[0].Segment.Name]++
			byGenre[e.Classifications[0].Genre.Name]++
		}

		nameLower := strings.ToLower(e.Name)
		nameCounts[nameLower]++

		if strings.Contains(nameLower, "parking") {
			parkingCount++
		}
		if strings.Contains(nameLower, "add-on") || strings.Contains(nameLower, "addon") ||
			strings.Contains(nameLower, "vip") || strings.Contains(nameLower, "package") {
			addonCount++
		}
	}

	fmt.Println("By Type:")
	for t, c := range byType {
		fmt.Printf("  %s: %d\n", t, c)
	}

	fmt.Println("\nBy Country (top 10):")
	for cc, c := range topN(byCountry, 10) {
		fmt.Printf("  %s: %d\n", cc, c)
	}

	fmt.Println("\nBy Segment:")
	for s, c := range bySegment {
		fmt.Printf("  %s: %d\n", s, c)
	}

	fmt.Println("\nBy Genre (top 10):")
	for g, c := range topN(byGenre, 10) {
		fmt.Printf("  %s: %d\n", g, c)
	}

	fmt.Println("\nPotential junk events:")
	fmt.Printf("  Parking: %d\n", parkingCount)
	fmt.Printf("  Add-ons/VIP: %d\n", addonCount)

	fmt.Println("\nDuplicate names (count > 1):")
	dupes := 0
	for name, c := range nameCounts {
		if c > 1 {
			fmt.Printf("  [%dx] %s\n", c, name)
			dupes += c - 1
		}
	}
	if dupes == 0 {
		fmt.Println("  (none)")
	} else {
		fmt.Printf("\n  Total duplicate entries: %d\n", dupes)
	}

	fmt.Println("\n--- Recommendations ---")
	if parkingCount > 0 {
		fmt.Printf("Consider filtering events with 'parking' in name (%d events)\n", parkingCount)
	}
	if addonCount > 0 {
		fmt.Printf("Consider filtering events with 'add-on', 'vip', 'package' in name (%d events)\n", addonCount)
	}
	if dupes > 0 {
		fmt.Printf("Consider deduplicating by event name (%d duplicates)\n", dupes)
	}
}

func filterParkingEvents(events []ticketmaster.Event) []ticketmaster.Event {
	var filtered []ticketmaster.Event
	for _, e := range events {
		nameLower := strings.ToLower(e.Name)
		if strings.Contains(nameLower, "parking") ||
			strings.Contains(nameLower, "add-on") ||
			strings.Contains(nameLower, "addon") ||
			strings.Contains(nameLower, "vip package") ||
			strings.Contains(nameLower, "hotel") ||
			strings.Contains(nameLower, "hospitality") {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

func filterDuplicateEvents(events []ticketmaster.Event) []ticketmaster.Event {
	seen := make(map[string]bool)
	var filtered []ticketmaster.Event
	for _, e := range events {
		key := strings.ToLower(e.Name) + "|" + e.Dates.Start.LocalDate
		if seen[key] {
			continue
		}
		seen[key] = true
		filtered = append(filtered, e)
	}
	return filtered
}

func topN(m map[string]int, n int) map[string]int {
	type kv struct {
		k string
		v int
	}
	var items []kv
	for k, v := range m {
		items = append(items, kv{k, v})
	}
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].v > items[i].v {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	result := make(map[string]int)
	for i := 0; i < n && i < len(items); i++ {
		result[items[i].k] = items[i].v
	}
	return result
}
