-- Voice profile questionnaire answers for AI Email Agent (write-as-user style).
ALTER TABLE email_agent
  ADD COLUMN VoiceProfileJSON LONGTEXT NULL AFTER ReplyPrompt;
