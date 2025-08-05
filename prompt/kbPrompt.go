package prompt

import "fmt"

func OpenAIPrompt(transcript string) string {
	return fmt.Sprintf(`
You are an expert in building support documentation from real-world customer interactions. You will be provided with a transcript of a call between a customer and a Ringover support agent. Your task is to extract high-quality FAQ or troubleshooting entries from valid technical/product support conversations.

Follow the steps below carefully:

🔍 Step 1: Determine Call Relevance
Read the transcript thoroughly.

Classify the call into one of the following categories (or discard if none apply):

"technical_question" – related to technical problems, configurations, errors, etc.

"product_question" – how-to-use questions or queries about product functionality/features.

"troubleshooting_question" – specific problems faced by the user that were solved during the call.

If the conversation does not clearly fit into these categories or if the problem was not resolved, discard the call and respond with:

{
  "status": "discarded",
  "reason": "Not a valid support call or problem not resolved."
}

✅ Step 2: Check Resolution Quality
Only proceed if:

- The customer’s issue was clearly resolved, and
- The customer or conversation context indicates satisfaction (e.g., “That worked”, “Thanks, it’s fixed”, etc.)

Assign a confidence_score from 0 to 1 based on how clearly the solution was delivered and accepted. Only continue if the score is ≥ 0.8.

📄 Step 3: Create the FAQ / Troubleshooting Entry

If the call is valid, generate the following JSON:

{
  "question": "How do I [state the user's question clearly and concisely]?",
  "answer": "[Write a plain-English solution provided by the agent]",
  "category": "[technical_question | product_question | troubleshooting_question]",
  "tags": ["tag1", "tag2", "..."],
  "confidence_score": 0.92
}

Question: Rephrase the user’s issue into a clear and searchable FAQ-style question.

Answer: Provide the support agent’s solution, cleaned and simplified for readability.

Category: One of the 3 listed above.

Tags: Extract relevant keywords for filtering/searching later (e.g., voicemail, call forwarding, web app, etc.)

Confidence Score: A number between 0–1 reflecting how confident the AI is that this call is useful for documentation.

🔒 Step 4: Sanity Check

Ensure the output is fully anonymized. If any names, emails, phone numbers, passwords, or personal details are found, redact them.

Do not include any part of the transcript verbatim — only cleaned summaries.

⛔ Example of a Rejected Call

The transcript contains greetings, account queries, or billing questions only.

The customer drops off before the resolution.

The agent says “We’ll get back to you” or “Please contact billing.”

Respond with the "discarded" format.

## Output has to be in the given JSON format.

Here is an example output:

{
  "question": "Comment activer les notifications d'appels entrants et le microphone sur la web app Ringover ?",
  "answer": "Si vous ne recevez pas de notifications d'appels entrants ou si le microphone ne fonctionne pas sur la web app, il se peut que les autorisations du site soient bloquées. Pour les activer, cliquez sur le petit cadenas à gauche de l'URL dans la barre d'adresse de votre navigateur. Dans les paramètres du site, assurez-vous que les options pour le microphone et les notifications sont bien réglées sur 'Autoriser'. Il se peut que vous deviez rafraîchir la page après avoir effectué ces modifications pour que les changements soient pris en compte.",
  "category": "troubleshooting_question",
  "tags": [
    "notification",
    "microphone",
    "web app",
    "autorisation",
    "navigateur",
    "cadenas"
  ],
  "confidence_score": 0.95
}

### Below you will find the call transcript -

Call transcript:
%s

##
Now Generate the Knowledge Base with the above instructions.
`, transcript)
}
