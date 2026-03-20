package ai

// All prompt templates are defined here as constants so they can be iterated on
// independently of the provider logic.

const promptSummarize = `You are a senior Go engineer reviewing a pull request's dependency impact report.
Given the following structured data, write a concise plain-English summary (3–5 sentences) that:
1. Describes what changed and whether it passed or failed policy.
2. Highlights the most important risk or concern (if any).
3. Mentions any notable binary size impact.
4. Ends with a one-sentence overall recommendation.

Do not repeat the raw numbers verbatim — synthesize them into natural language.
Do not use bullet points. Output only the paragraph, no headings.

DATA:
%s`

const promptEnhanceRisk = `You are a Go dependency security and maintainability expert.
A new direct dependency has been added to a Go project. 
Evaluate its risk on a scale of 1–10 (1=trivial, 10=extreme risk).
Consider: maintenance status, CVE history, ecosystem reputation, transitive bloat, and license risk.

Return your response in this exact JSON format:
{"score": <integer 1-10>, "reasons": ["<reason1>", "<reason2>", ...]}

Module: %s
Version: %s
Transitive packages added: %d
Transitive modules added: %d
Initial static score: %d
Static reasons: %v`

const promptValidateReason = `You are a Go architect reviewing a PR where a developer added a new direct dependency.
Evaluate whether the stated reason is a good justification for this dependency choice.
Consider: could a stdlib package suffice? Is the module overkill for the stated purpose? Are there well-known lighter alternatives?

Return a JSON array of concern strings (empty array if no concerns):
["<concern1>", "<concern2>"]

Module added: %s
Developer's stated reason: %s
Relevant PR diff (truncated):
%s`

const promptSuggestAlternatives = `You are a Go dependency optimization expert.
For each of the following newly added Go dependencies that appear heavy or large, suggest a lighter alternative if one exists.
Only suggest an alternative if you are confident one exists and is production-ready.
If no good alternative exists, omit that module from your response.

Return a JSON object mapping module paths to suggestion strings:
{"github.com/heavy/module": "Consider 'github.com/lighter/module' — it provides the same feature with X fewer transitive packages."}

Modules to evaluate:
%s`

const promptAnalyzeTrend = `You are a Go platform engineering expert analyzing a project's dependency and binary size growth over time.
Identify any concerning trends, anomalies, or patterns in the data.
Be specific about which time periods show unusual growth.
Write 2–4 sentences of plain-English analysis followed by a bullet-point list of specific findings.

History data (JSON):
%s`

const promptGenerateConfig = `You are a Go CI expert. Generate a .godepsguard.yaml config file based on the user's plain-English description of their policy requirements.

Rules:
- Only include fields that are directly implied by the description.
- Use sensible numeric defaults if the user states a concept but not a number.
- Blocked licenses should use SPDX identifiers (e.g., GPL-3.0, AGPL-3.0).
- Output ONLY the raw YAML, no markdown fences, no explanation.

User description: %s`
