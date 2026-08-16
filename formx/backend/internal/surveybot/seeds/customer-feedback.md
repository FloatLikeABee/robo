---
slug: customer-feedback
title: Customer feedback survey
tags: [feedback, customer, csat]
---

# Instructions
Collect concise feedback. Use the selector for satisfaction score.

## Q1 — Customer name
- field: customer_name
- collect: text
- required: true
- prompt: What is the customer's name?

## Q2 — Satisfaction
- field: satisfaction
- collect: mcp_html
- widget: select
- options: [Very satisfied, Satisfied, Neutral, Dissatisfied, Very dissatisfied]
- required: true
- prompt: How satisfied were they overall?

## Q3 — What went well
- field: went_well
- collect: text
- required: false
- prompt: What went well?

## Q4 — Improvements
- field: improvements
- collect: text
- required: false
- prompt: What could we improve?
