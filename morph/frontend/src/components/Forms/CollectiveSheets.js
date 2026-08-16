import React, { useEffect, useMemo, useState, useCallback, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import { getFormAssignments, getFormTemplate, createFormAnswer } from '../../services/formsApi';
import './Forms.css';

export default function CollectiveSheets({ actor }) {
  const [searchParams, setSearchParams] = useSearchParams();
  const openedFromQueryRef = useRef(null);
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [activeAssignment, setActiveAssignment] = useState(null);
  const [fields, setFields] = useState([]);
  const [answers, setAnswers] = useState({});
  const [alert, setAlert] = useState(null);

  const showAlert = useCallback((message, type) => {
    setAlert({ message, type });
    setTimeout(() => setAlert(null), 5000);
  }, []);

  const loadAssignments = async () => {
    try {
      setLoading(true);
      const filters = {};
      if (actor?.role !== 'admin') {
        filters.assignee_user_id = actor?.user_id || '';
        filters.assignee_user_type = actor?.user_type || 'staff';
      }
      const list = await getFormAssignments(filters);
      setRows(Array.isArray(list) ? list : []);
    } catch (error) {
      showAlert('Error loading collective sheets: ' + (error.response?.data?.error || error.message), 'error');
      setRows([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAssignments();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [actor?.user_id, actor?.user_type, actor?.role]);

  const pendingRows = useMemo(
    () => rows.filter(row => row.status === 'pending'),
    [rows]
  );

  const completedRows = useMemo(
    () => rows.filter(row => row.status === 'completed'),
    [rows]
  );

  const openSubmitModal = useCallback(async (assignment) => {
    try {
      const form = await getFormTemplate(assignment.form_id);
      const formFields = Array.isArray(form?.fields) ? form.fields : [];
      const initial = {};
      formFields.forEach((field) => {
        initial[field.name] = '';
      });
      setFields(formFields);
      setAnswers(initial);
      setActiveAssignment(assignment);
      setShowModal(true);
    } catch (error) {
      showAlert('Error loading sheet fields: ' + (error.response?.data?.error || error.message), 'error');
    }
  }, [showAlert]);

  const clearOpenAssignmentParam = useCallback(() => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.delete('openAssignment');
        return next;
      },
      { replace: true }
    );
  }, [setSearchParams]);

  useEffect(() => {
    const rawId = searchParams.get('openAssignment');
    if (!rawId || loading) return;
    if (openedFromQueryRef.current === rawId) return;
    const row = rows.find((r) => String(r.id) === String(rawId) && r.status === 'pending');
    if (!row) {
      if (!loading) {
        clearOpenAssignmentParam();
      }
      return;
    }
    openedFromQueryRef.current = rawId;
    void openSubmitModal(row).finally(() => {
      clearOpenAssignmentParam();
    });
  }, [loading, rows, searchParams, openSubmitModal, clearOpenAssignmentParam]);

  useEffect(() => {
    if (!searchParams.get('openAssignment')) {
      openedFromQueryRef.current = null;
    }
  }, [searchParams]);

  const closeModal = () => {
    setShowModal(false);
    setActiveAssignment(null);
    setFields([]);
    setAnswers({});
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!activeAssignment) return;
    try {
      await createFormAnswer(
        {
          form_id: activeAssignment.form_id,
          user_id: String(actor?.user_id || '1'),
          user_type: actor?.user_type || activeAssignment.form_user_type || 'staff',
          assignment_id: activeAssignment.id,
          answers,
        },
        actor
      );
      showAlert('Sheet submitted successfully', 'success');
      closeModal();
      loadAssignments();
    } catch (error) {
      showAlert('Error submitting sheet: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  return (
    <div className="form-answers-container">
      <div className="form-answers-content">
        {alert && (
          <div className={`alert alert-${alert.type}`}>
            {alert.message}
          </div>
        )}
        <div className="toolbar">
          <div className="filter-group">
            <strong>Collective Sheets</strong>
            <span>Pending: {pendingRows.length}</span>
            <span>Completed: {completedRows.length}</span>
          </div>
        </div>
        {loading ? (
          <div className="loading">Loading collective sheets...</div>
        ) : (
          <div className="forms-scroll">
            <div className="forms-grid">
              {pendingRows.map((item) => (
                <div key={item.id} className="form-card">
                  <div className="form-card-body">
                    <h3>{item.form_name}</h3>
                    <div className="meta">
                      <strong>Assigned:</strong> {item.assigned_at ? new Date(item.assigned_at).toLocaleString() : 'N/A'}
                    </div>
                  </div>
                  <div className="actions">
                    <button
                      type="button"
                      className="btn btn-small"
                      onClick={() => openSubmitModal(item)}
                    >
                      Submit now
                    </button>
                  </div>
                </div>
              ))}
              {pendingRows.length === 0 && (
                <div className="empty-state">
                  <h3>No pending sheets</h3>
                  <p>Assigned sheets will appear here until submitted.</p>
                </div>
              )}
            </div>
            {completedRows.length > 0 && (
              <div style={{ marginTop: '1.5rem' }}>
                <h3 style={{ marginBottom: '0.75rem' }}>Completed</h3>
                <div className="forms-grid">
                  {completedRows.map((item) => (
                    <div key={item.id} className="form-card">
                      <div className="form-card-body">
                        <h3>{item.form_name}</h3>
                        <div className="meta">
                          <strong>Completed:</strong> {item.completed_at ? new Date(item.completed_at).toLocaleString() : 'N/A'}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {showModal && activeAssignment && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content modal-content--quick-sheet" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Submit: {activeAssignment.form_name}</h2>
              <button className="close-btn" onClick={closeModal}>&times;</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="answers-editor">
                {fields.map((field, index) => (
                  <div key={index} className="answer-field-item">
                    <label>{field.label} {field.required ? '*' : ''}</label>
                    <input
                      type="text"
                      value={answers[field.name] || ''}
                      onChange={(e) => setAnswers((prev) => ({ ...prev, [field.name]: e.target.value }))}
                      placeholder={field.placeholder || ''}
                      required={field.required}
                    />
                  </div>
                ))}
              </div>
              <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={closeModal}>
                  Cancel
                </button>
                <button type="submit" className="btn">
                  Submit sheet
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
