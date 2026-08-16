import { useCallback, useState } from 'react';
import {
  EventInfoSubmitCard,
  defaultEventInfoTimeLocal,
  type EventInfoSubmitValues,
} from '../components/EventInfoSubmitCard';
import { api } from '../lib/api';

export function PublicEventInfoSubmit() {
  const [values, setValues] = useState<EventInfoSubmitValues>(() => ({
    title: '',
    detail: '',
    reporter: '',
    time: defaultEventInfoTimeLocal(),
  }));
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const patch = useCallback((patch: Partial<EventInfoSubmitValues>) => {
    setValues((prev) => ({ ...prev, ...patch }));
  }, []);

  const submit = async () => {
    if (!values.title.trim()) {
      setError('Title is required');
      return;
    }
    const t = values.time ? new Date(values.time).toISOString() : new Date().toISOString();
    setSubmitting(true);
    setError(null);
    try {
      await api.public.createEventInfo({
        title: values.title.trim(),
        detail: values.detail,
        reporter: values.reporter.trim(),
        time: t,
      });
      setSuccess('Thank you — your entry was submitted.');
      setValues({
        title: '',
        detail: '',
        reporter: '',
        time: defaultEventInfoTimeLocal(),
      });
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Submit failed');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col items-center justify-center p-4 sm:p-8">
      <div className="w-full max-w-2xl mb-6 text-center">
        <p className="text-xs uppercase tracking-[0.28em] text-violet-400/90 font-medium">SurveyX</p>
        <h1 className="text-2xl sm:text-3xl font-semibold text-white mt-2">Events &amp; Info</h1>
        <p className="text-sm text-slate-400 mt-2">Submit an operational note for your team.</p>
      </div>
      <EventInfoSubmitCard
        variant="public"
        title="New entry"
        subtitle="Fill in the details below. Title is required."
        values={values}
        onChange={patch}
        onSubmit={submit}
        submitting={submitting}
        error={error}
        successMessage={success}
        showCancel={false}
      />
    </div>
  );
}
