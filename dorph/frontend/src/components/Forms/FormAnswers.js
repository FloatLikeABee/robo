import React, { useState, useEffect, useMemo } from 'react';
import {
  getFormAnswers,
  createFormAnswer,
  deleteFormAnswer,
  getFormTemplates,
  getFormTemplate
} from '../../services/formsApi';
import { useConfirm } from '../ConfirmDialog';
import './FormAnswers.css';
import './Forms.css';

const FormAnswers = ({ actor }) => {
  const { confirm } = useConfirm();
  const [answers, setAnswers] = useState([]);
  const [filteredAnswers, setFilteredAnswers] = useState([]);
  const [forms, setForms] = useState([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [detailAnswer, setDetailAnswer] = useState(null);
  const pageSize = 10;
  const [formFilter, setFormFilter] = useState('');
  const [userTypeFilter, setUserTypeFilter] = useState('');
  const [userIdFilter, setUserIdFilter] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [alert, setAlert] = useState(null);
  const [answerData, setAnswerData] = useState({
    form_id: '',
    user_id: '',
    user_type: '',
    answers: {}
  });
  const [selectedFormFields, setSelectedFormFields] = useState([]);

  useEffect(() => {
    loadForms();
    loadAnswers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    filterAnswers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [answers, formFilter, userTypeFilter, userIdFilter]);

  useEffect(() => {
    setPage(1);
  }, [formFilter, userTypeFilter, userIdFilter]);

  const totalPages = useMemo(() => {
    const n = filteredAnswers.length;
    return Math.max(1, Math.ceil(n / pageSize) || 1);
  }, [filteredAnswers.length, pageSize]);

  const safePage = Math.min(page, totalPages);
  const pageRows = useMemo(() => {
    const start = (safePage - 1) * pageSize;
    return filteredAnswers.slice(start, start + pageSize);
  }, [filteredAnswers, safePage, pageSize]);

  useEffect(() => {
    if (page > totalPages) setPage(totalPages);
  }, [page, totalPages]);

  const loadForms = async () => {
    try {
      const data = await getFormTemplates();
      setForms(Array.isArray(data) ? data : []);
    } catch (error) {
      showAlert('Error loading sheets: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const loadAnswers = async () => {
    try {
      setLoading(true);
      const data = await getFormAnswers();
      // Ensure we always work with an array to avoid null/undefined issues
      setAnswers(Array.isArray(data) ? data : (data ? [data] : []));
    } catch (error) {
      showAlert('Error loading responses: ' + (error.response?.data?.error || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const filterAnswers = () => {
    // Start from a safe array value
    let filtered = Array.isArray(answers) ? answers : [];
    
    if (formFilter) {
      filtered = filtered.filter(a => a.form_id === formFilter);
    }
    if (userTypeFilter) {
      filtered = filtered.filter(a => a.user_type === userTypeFilter);
    }
    if (userIdFilter) {
      const q = userIdFilter.toLowerCase();
      filtered = filtered.filter(a =>
        String(a.user_id ?? '').toLowerCase().includes(q)
      );
    }
    
    // Make sure filteredAnswers is never null
    setFilteredAnswers(Array.isArray(filtered) ? filtered : []);
  };

  const showAlert = (message, type) => {
    setAlert({ message, type });
    setTimeout(() => setAlert(null), 5000);
  };

  const loadFormFields = async (formId) => {
    if (!formId) {
      setSelectedFormFields([]);
      return;
    }

    try {
      const form = await getFormTemplate(formId);
      const fields = Array.isArray(form?.fields) ? form.fields : [];
      setSelectedFormFields(fields);
      
      // Initialize answer data with empty values for each field
      const initialAnswers = {};
      fields.forEach(field => {
        initialAnswers[field.name] = '';
      });
      setAnswerData(prev => ({
        ...prev,
        answers: { ...prev.answers, ...initialAnswers }
      }));
    } catch (error) {
      showAlert('Error loading sheet fields: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const openCreateModal = () => {
    setAnswerData({
      form_id: '',
      user_id: String(actor?.user_id || ''),
      user_type: actor?.user_type || '',
      answers: {}
    });
    setSelectedFormFields([]);
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setAnswerData({
      form_id: '',
      user_id: '',
      user_type: '',
      answers: {}
    });
    setSelectedFormFields([]);
  };

  const handleFormChange = async (formId) => {
    setAnswerData(prev => ({ ...prev, form_id: formId, answers: {} }));
    await loadFormFields(formId);
  };

  const handleAnswerChange = (fieldName, value) => {
    setAnswerData(prev => ({
      ...prev,
      answers: {
        ...prev.answers,
        [fieldName]: value
      }
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validate
    if (!answerData.form_id) {
      showAlert('Sheet template is required', 'error');
      return;
    }
    if (!answerData.user_id.trim()) {
      showAlert('User ID is required', 'error');
      return;
    }
    if (!answerData.user_type) {
      showAlert('User type is required', 'error');
      return;
    }

    try {
      await createFormAnswer(answerData);
      showAlert('Response submitted successfully!', 'success');
      
      closeModal();
      loadAnswers();
    } catch (error) {
      showAlert('Error saving response: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const handleDelete = async (id) => {
    const ok = await confirm({
      title: 'Delete response',
      message: 'Are you sure you want to delete this response?',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    
    try {
      await deleteFormAnswer(id);
      showAlert('Response deleted successfully!', 'success');
      loadAnswers();
    } catch (error) {
      showAlert('Error deleting response: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return dateString;
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
            <label htmlFor="formFilter">Sheet:</label>
            <select
              id="formFilter"
              value={formFilter}
              onChange={(e) => setFormFilter(e.target.value)}
            >
              <option value="">All sheets</option>
              {(forms || []).map(form => (
                <option key={form.id} value={form.id}>{form.name}</option>
              ))}
            </select>
            <label htmlFor="userTypeFilter">User Type:</label>
            <select
              id="userTypeFilter"
              value={userTypeFilter}
              onChange={(e) => setUserTypeFilter(e.target.value)}
            >
              <option value="">All Types</option>
              <option value="student">Member</option>
              <option value="staff">Employee</option>
            </select>
            <label htmlFor="userIdFilter">User ID:</label>
            <input
              type="text"
              id="userIdFilter"
              value={userIdFilter}
              onChange={(e) => setUserIdFilter(e.target.value)}
              placeholder="Filter by user ID"
              style={{
                padding: '0.5rem 1rem',
                background: 'var(--chat-msg-bg, #101820)',
                color: 'var(--chat-text, #ece8f7)',
                border: '1px solid var(--chat-border, rgba(34, 211, 238, 0.18))',
                borderRadius: '4px',
                fontSize: '1rem'
              }}
            />
          </div>
          <button className="btn" onClick={openCreateModal}>
            + Submit New Response
          </button>
        </div>

        {loading ? (
          <div className="loading">Loading responses...</div>
        ) : filteredAnswers.length === 0 ? (
          <div className="empty-state">
            <h3>No responses found</h3>
            <p>Submit your first response to get started</p>
          </div>
        ) : (
          <>
            <div className="answers-table-panel">
              <div className="answers-table-scroll">
                <table className="answers-table">
                  <thead>
                    <tr>
                      <th>Sheet name</th>
                      <th>User ID</th>
                      <th>User Type</th>
                      <th>Submitted At</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {pageRows.map(answer => (
                      <tr key={answer.id}>
                        <td>{answer.form_name || 'Unknown'}</td>
                        <td>{answer.user_id}</td>
                        <td>
                          <span className={`badge badge-${answer.user_type}`}>
                            {answer.user_type}
                          </span>
                        </td>
                        <td>{formatDate(answer.submitted_at)}</td>
                        <td className="answers-actions-cell">
                          <button
                            type="button"
                            className="btn btn-secondary btn-small"
                            onClick={() => setDetailAnswer(answer)}
                          >
                            Detail
                          </button>
                          <button
                            type="button"
                            className="btn btn-danger btn-small btn-icon-delete"
                            onClick={() => handleDelete(answer.id)}
                            title="Delete response"
                            aria-label="Delete response"
                          >
                            <svg viewBox="0 0 24 24" width="15" height="15" aria-hidden="true" focusable="false">
                              <path d="M3 6h18" />
                              <path d="M8 6V4h8v2" />
                              <path d="M19 6l-1 14H6L5 6" />
                              <path d="M10 11v6" />
                              <path d="M14 11v6" />
                            </svg>
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="answers-pagination">
              <button
                type="button"
                className="btn btn-secondary btn-small"
                disabled={safePage <= 1}
                onClick={() => setPage(p => Math.max(1, p - 1))}
              >
                Previous
              </button>
              <span className="answers-pagination-info">
                Page {safePage} of {totalPages}
                <span className="answers-pagination-count">
                  ({filteredAnswers.length} total)
                </span>
              </span>
              <button
                type="button"
                className="btn btn-secondary btn-small"
                disabled={safePage >= totalPages}
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              >
                Next
              </button>
            </div>
          </>
        )}
      </div>

      {/* Response detail (read-only) */}
      {detailAnswer && (
        <div className="modal-overlay" onClick={() => setDetailAnswer(null)}>
          <div className="modal-content modal-content-detail" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Response details</h2>
              <button type="button" className="close-btn" onClick={() => setDetailAnswer(null)}>&times;</button>
            </div>
            <div className="response-detail-meta">
              <p>
                <strong>Sheet:</strong> {detailAnswer.form_name || 'Unknown'}
              </p>
              <p>
                <strong>User ID:</strong> {detailAnswer.user_id}
                {' · '}
                <span className={`badge badge-${detailAnswer.user_type}`}>{detailAnswer.user_type}</span>
              </p>
              <p>
                <strong>Submitted:</strong> {formatDate(detailAnswer.submitted_at)}
              </p>
            </div>
            <div className="response-detail-fields">
              <h3 className="response-detail-fields-title">Responses</h3>
              <div className="response-detail-scroll">
                {Object.entries(detailAnswer.answers || {}).map(([key, value]) => (
                  <div key={key} className="answer-item">
                    <strong>{key}</strong>
                    <span>{String(value)}</span>
                  </div>
                ))}
              </div>
            </div>
            <div className="modal-actions">
              <button type="button" className="btn btn-secondary" onClick={() => setDetailAnswer(null)}>
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create / Edit modal */}
      {showModal && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Submit New Response</h2>
              <button className="close-btn" onClick={closeModal}>&times;</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label htmlFor="answerFormID">Sheet Template *</label>
                <select
                  id="answerFormID"
                  value={answerData.form_id}
                  onChange={(e) => handleFormChange(e.target.value)}
                  required
                >
                  <option value="">Select a sheet...</option>
                  {(forms || []).map(form => (
                    <option key={form.id} value={form.id}>{form.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label htmlFor="answerUserID">User ID *</label>
                <input
                  type="text"
                  id="answerUserID"
                  value={answerData.user_id}
                  onChange={(e) => setAnswerData({ ...answerData, user_id: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label htmlFor="answerUserType">User Type *</label>
                <select
                  id="answerUserType"
                  value={answerData.user_type}
                  onChange={(e) => setAnswerData({ ...answerData, user_type: e.target.value })}
                  required
                >
                  <option value="">Select...</option>
                  <option value="student">Student</option>
                  <option value="staff">Employee</option>
                </select>
              </div>
              <div className="form-group">
                <label>Responses</label>
                <div className="answers-editor">
                  {selectedFormFields.length === 0 ? (
                    <p style={{ color: 'var(--chat-text-muted, #8899b0)' }}>Select a sheet template to load fields</p>
                  ) : (
                    selectedFormFields.map((field, index) => (
                      <div key={index} className="answer-field-item">
                        <label>
                          {field.label} {field.required && <span style={{ color: '#dc3545' }}>*</span>}
                        </label>
                        <input
                          type="text"
                          value={answerData.answers[field.name] || ''}
                          onChange={(e) => handleAnswerChange(field.name, e.target.value)}
                          placeholder={field.placeholder || ''}
                          required={field.required}
                        />
                      </div>
                    ))
                  )}
                </div>
              </div>
              <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={closeModal}>
                  Cancel
                </button>
                <button type="submit" className="btn">
                  Save Response
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default FormAnswers;
