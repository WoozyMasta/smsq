// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/woozymasta/smsq/internal/keys"
	slist "github.com/woozymasta/steamserverlist"
)

// filterKeyHelpFooter renders the accepted filter catalog from its validation registry.
func filterKeyHelpFooter() string {
	booleanNames := make([]string, 0, len(keys.Registry))
	valueNames := make([]string, 0, len(keys.Registry))

	expressionWidth := 0
	for name := range keys.Registry {
		info := keys.Registry[name]
		if info.Flag {
			booleanNames = append(booleanNames, name)
		} else {
			valueNames = append(valueNames, name)
		}

		if width := len(name) + 1 + len(info.Example); width > expressionWidth {
			expressionWidth = width
		}
	}

	sort.Strings(booleanNames)
	sort.Strings(valueNames)

	var footer strings.Builder

	footer.WriteString("Filters:\n")
	footer.WriteString("  Use the listed keys with --filter, --nor, and --nand.\n")
	footer.WriteByte('\n')
	footer.WriteString("  Flag filters:\n")
	footer.WriteString("    Binary flags accept 0/1, true/false, yes/no, or y/n.\n")

	for _, name := range booleanNames {
		info := keys.Registry[name]
		fmt.Fprintf(&footer, "    %-*s  %s\n", expressionWidth, name+"="+info.Example, info.Desc)
	}

	footer.WriteString("\n  Value filters:\n")
	for _, name := range valueNames {
		info := keys.Registry[name]
		fmt.Fprintf(&footer, "    %-*s  %s\n", expressionWidth, name+"="+info.Example, info.Desc)
	}

	return footer.String()
}

// filterKeyMarkdownFooter renders the accepted filter catalog for generated documentation.
func filterKeyMarkdownFooter() string {
	booleanNames := make([]string, 0, len(keys.Registry))
	valueNames := make([]string, 0, len(keys.Registry))

	for name, info := range keys.Registry {
		if info.Flag {
			booleanNames = append(booleanNames, name)
		} else {
			valueNames = append(valueNames, name)
		}
	}

	sort.Strings(booleanNames)
	sort.Strings(valueNames)

	var footer strings.Builder
	footer.WriteString("## Filter keys\n\n")
	footer.WriteString("Use these keys with `--filter`, `--nor`, and `--nand`. " +
		"Filter behavior may vary by protocol and game.\n\n")
	footer.WriteString("### Flag filters\n\n")
	footer.WriteString("Binary flag operands accept `0`/`1`, `true`/`false`, `yes`/`no`, or `y`/`n`.\n\n")

	for _, name := range booleanNames {
		info := keys.Registry[name]
		fmt.Fprintf(&footer, "* `%s=%s`: %s\n", name, info.Example, info.Desc)
	}

	footer.WriteString("\n### Value filters\n\n")
	for _, name := range valueNames {
		info := keys.Registry[name]
		fmt.Fprintf(&footer, "* `%s=%s`: %s\n", name, info.Example, info.Desc)
	}

	return footer.String()
}

// buildFilter parses shared CLI predicates and validates the Steam filter groups.
func buildFilter(filters filterFlags) (slist.QueryFilter, string, uint32, error) {
	if filters.RawFilter != "" {
		appID := filters.AppID
		if strings.HasPrefix(filters.RawFilter, `\appid\`) {
			rawAppID, err := rawFilterAppID(filters.RawFilter)
			if err != nil {
				return slist.QueryFilter{}, "", 0, err
			}

			appID = rawAppID
		} else if appID > 0 {
			filters.RawFilter = fmt.Sprintf(`\appid\%d%s`, appID, filters.RawFilter)
		}

		if appID == 0 {
			return slist.QueryFilter{}, "", 0, errors.New(
				"provide --app-id or include a positive AppID in --raw-filter",
			)
		}

		return slist.QueryFilter{}, filters.RawFilter, appID, nil
	}
	if filters.AppID == 0 {
		return slist.QueryFilter{}, "", 0, errors.New("--app-id must be a positive integer")
	}

	positive, err := keys.Parse(filters.Filter)
	if err != nil {
		return slist.QueryFilter{}, "", 0, err
	}
	nor, err := keys.Parse(filters.NOR)
	if err != nil {
		return slist.QueryFilter{}, "", 0, err
	}
	nand, err := keys.Parse(filters.NAND)
	if err != nil {
		return slist.QueryFilter{}, "", 0, err
	}

	conditions := make([]slist.Predicate, 0, len(positive)+1)
	conditions = append(conditions, slist.Predicate{
		Key:   slist.KeyAppID,
		Value: strconv.FormatUint(uint64(filters.AppID), 10),
	})
	conditions = append(conditions, convertPredicates(positive)...)

	var group *slist.PredicateGroup
	if len(nor) > 0 {
		group = &slist.PredicateGroup{
			Kind:       slist.GroupNOR,
			Predicates: convertPredicates(nor),
		}
	}
	if len(nand) > 0 {
		group = &slist.PredicateGroup{
			Kind:       slist.GroupNAND,
			Predicates: convertPredicates(nand),
		}
	}

	filter, err := slist.NewQueryFilter(conditions, group)
	if err != nil {
		return slist.QueryFilter{}, "", 0, fmt.Errorf("validate filter: %w", err)
	}

	return filter, filter.String(), filters.AppID, nil
}

// rawFilterAppID reads the leading AppID pair from a complete raw expression.
func rawFilterAppID(filter string) (uint32, error) {
	value, _, _ := strings.Cut(strings.TrimPrefix(filter, `\appid\`), `\`)
	appID, err := strconv.ParseUint(value, 10, 32)
	if err != nil || appID == 0 {
		return 0, errors.New("raw filter must start with \\appid\\<positive id>")
	}

	return uint32(appID), nil
}

// convertPredicates maps parsed CLI predicates to the Web API filter contract.
func convertPredicates(predicates []keys.Predicate) []slist.Predicate {
	converted := make([]slist.Predicate, len(predicates))

	for index, predicate := range predicates {
		converted[index] = slist.Predicate{
			Key: slist.FilterKey(predicate.Key), Value: predicate.Value,
		}
	}

	return converted
}
