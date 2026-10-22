import React, { useState, useEffect } from 'react';
import {
  getFormTemplates,
  getFormTemplate,
  createFormTemplate,
  updateFormTemplate,
  deleteFormTemplate,
  getAssignableUsers,
  createFormAssignments,
} from '../../services/formsApi';
import { useConfirm } from '../ConfirmDialog';
import './Forms.css';

const Forms = ({ actor }) => {
  const { confirm } = useConfirm();
  const [forms, setForms] = useState([]);
  const [filteredForms, setFilteredForms] = useState([]);
  const [loading, setLoading] = useState(true);
  const [userTypeFilter, setUserTypeFilter] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editingForm, setEditingForm] = useState(null);
  const [alert, setAlert] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    user_type: '',
    fields: []
  });
  const [assigningForm, setAssigningForm] = useState(null);
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [assignSearch, setAssignSearch] = useState('');
  const [assigneeOptions, setAssigneeOptions] = useState([]);
  const [assigneeLoading, setAssigneeLoading] = useState(false);
  const [selectedAssigneeKeys, setSelectedAssigneeKeys] = useState({});

  useEffect(() => {
    loadForms();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    filterForms();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [forms, userTypeFilter]);

  const loadForms = async () => {
    try {
      setLoading(true);
      const data = await getFormTemplates();
      setForms(Array.isArray(data) ? data : []);
    } catch (error) {
      showAlert('Error loading sheets: ' + (error.response?.data?.error || error.message), 'error');
      setForms([]);
    } finally {
      setLoading(false);
    }
  };

  const filterForms = () => {
    const list = Array.isArray(forms) ? forms : [];
    if (!userTypeFilter) {
      setFilteredForms(list);
    } else {
      setFilteredForms(list.filter(f => f.user_type === userTypeFilter));
    }
  };

  const showAlert = (message, type) => {
    setAlert({ message, type });
    setTimeout(() => setAlert(null), 5000);
  };

  const openCreateModal = () => {
    setEditingForm(null);
    setFormData({
      name: '',
      description: '',
      user_type: '',
      fields: []
    });
    setShowModal(true);
  };

  const openEditModal = async (id) => {
    try {
      const form = await getFormTemplate(id);
      setEditingForm(id);
      setFormData({
        name: form.name,
        description: form.description || '',
        user_type: form.user_type,
        fields: (form.fields || []).map((f) => ({
          name: f.name || '',
          label: f.label || '',
          type: 'text',
          required: !!f.required,
          placeholder: f.placeholder || '',
        })),
      });
      setShowModal(true);
    } catch (error) {
      showAlert('Error loading sheet: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingForm(null);
    setFormData({
      name: '',
      description: '',
      user_type: '',
      fields: []
    });
  };

  const addField = () => {
    setFormData({
      ...formData,
      fields: [...formData.fields, {
        name: '',
        label: '',
        type: 'text',
        required: false,
        placeholder: '',
      }]
    });
  };

  const updateField = (index, field) => {
    const newFields = [...formData.fields];
    newFields[index] = { ...newFields[index], ...field };
    setFormData({ ...formData, fields: newFields });
  };

  const removeField = (index) => {
    const newFields = formData.fields.filter((_, i) => i !== index);
    setFormData({ ...formData, fields: newFields });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validate
    if (!formData.name.trim()) {
      showAlert('Sheet name is required', 'error');
      return;
    }
    if (!formData.user_type) {
      showAlert('User type is required', 'error');
      return;
    }

    // Filter out empty fields; Quick Sheets only support text questions
    const validFields = formData.fields
      .filter(f => f.name && f.label)
      .map(f => ({
        name: f.name,
        label: f.label,
        type: 'text',
        required: !!f.required,
        placeholder: f.placeholder || '',
      }));

    try {
      const templateData = {
        name: formData.name.trim(),
        description: formData.description.trim(),
        user_type: formData.user_type,
        fields: validFields
      };

      if (editingForm) {
        await updateFormTemplate(editingForm, templateData);
        showAlert('Sheet updated successfully!', 'success');
      } else {
        await createFormTemplate(templateData);
        showAlert('Sheet created successfully!', 'success');
      }
      
      closeModal();
      loadForms();
    } catch (error) {
      showAlert('Error saving sheet: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const handleDelete = async (id) => {
    const ok = await confirm({
      title: 'Delete sheet',
      message: 'Are you sure you want to delete this sheet?',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    
    try {
      await deleteFormTemplate(id);
      showAlert('Sheet deleted successfully!', 'success');
      loadForms();
    } catch (error) {
      showAlert('Error deleting sheet: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const loadAssigneeOptions = async (userType, query = '') => {
    try {
      setAssigneeLoading(true);
      const list = await getAssignableUsers(userType, query);
      setAssigneeOptions(Array.isArray(list) ? list : []);
    } catch (error) {
      showAlert('Error loading assignees: ' + (error.response?.data?.error || error.message), 'error');
      setAssigneeOptions([]);
    } finally {
      setAssigneeLoading(false);
    }
  };

  const openAssignModal = async (form) => {
    setAssigningForm(form);
    setAssignSearch('');
    setSelectedAssigneeKeys({});
    setShowAssignModal(true);
    await loadAssigneeOptions(form.user_type || 'student');
  };

  const closeAssignModal = () => {
    setAssigningForm(null);
    setAssignSearch('');
    setAssigneeOptions([]);
    setSelectedAssigneeKeys({});
    setShowAssignModal(false);
  };

  const assigneeKey = (row) => `${row.user_type}:${row.user_id}`;

  const toggleAssignee = (row) => {
    const key = assigneeKey(row);
    setSelectedAssigneeKeys(prev => ({ ...prev, [key]: !prev[key] }));
  };

  const assignToSelf = () => {
    if (!assigningForm) return;
    const uid = String(actor?.user_id || '').trim();
    if (!uid) {
      showAlert('Your account is missing a user id; sign in again.', 'error');
      return;
    }
    const me = {
      user_id: uid,
      user_type: assigningForm.user_type || 'staff',
      name: String(actor?.name || 'Current User'),
      email: '',
    };
    const key = assigneeKey(me);
    setAssigneeOptions(prev => {
      const exists = prev.some(item => assigneeKey(item) === key);
      return exists ? prev : [me, ...prev];
    });
    setSelectedAssigneeKeys(prev => ({ ...prev, [key]: true }));
  };

  const filteredAssigneeOptions = assigneeOptions.filter(row => {
    if (!assignSearch.trim()) return true;
    const q = assignSearch.trim().toLowerCase();
    const blob = `${row.user_id} ${row.name || ''} ${row.email || ''}`.toLowerCase();
    return blob.includes(q);
  });

  const submitAssignments = async () => {
    if (!assigningForm) return;
    const selected = assigneeOptions.filter(row => selectedAssigneeKeys[assigneeKey(row)]);
    if (selected.length === 0) {
      showAlert('Select at least one assignee', 'error');
      return;
    }
    try {
      await createFormAssignments({
        form_id: assigningForm.id,
        assignees: selected.map(row => ({
          user_id: String(row.user_id),
          user_type: row.user_type,
          user_name: row.name || '',
        })),
      }, actor);
      showAlert(`Assigned "${assigningForm.name}" to ${selected.length} user(s)`, 'success');
      closeAssignModal();
    } catch (error) {
      showAlert('Error assigning sheet: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return dateString;
    }
  };

  return (
    <div className="forms-container">
      <div className="forms-content">
        {alert && (
          <div className={`alert alert-${alert.type}`}>
            {alert.message}
          </div>
        )}

        <div className="toolbar">
          <div className="filter-group">
            <label htmlFor="userTypeFilter">Filter:</label>
            <select
              id="userTypeFilter"
              value={userTypeFilter}
              onChange={(e) => setUserTypeFilter(e.target.value)}
            >
              <option value="">All Types</option>
              <option value="student">Member sheets</option>
              <option value="staff">Employee sheets</option>
            </select>
          </div>
          <button className="btn" onClick={openCreateModal}>
            + Create New Sheet
          </button>
        </div>

        {loading ? (
          <div className="loading">Loading sheets...</div>
        ) : (filteredForms || []).length === 0 ? (
          <div className="empty-state">
            <h3>No sheets found</h3>
            <p>Create your first sheet template to get started</p>
          </div>
        ) : (
          <div className="forms-scroll">
            <div className="forms-grid">
              {(filteredForms || []).map(form => (
                <div key={form.id} className="form-card">
                  <div className="form-card-body">
                    <h3>{form.name}</h3>
                    <div className="meta">
                      <strong>Type:</strong> {form.user_type}
                      {' · '}
                      <strong>Fields:</strong> {form.fields?.length || 0}
                      {' · '}
                      <strong>Created:</strong> {formatDate(form.created_at)}
                    </div>
                    {form.description && (
                      <p className="form-card-desc">{form.description}</p>
                    )}
                  </div>
                  <div className="actions">
                    <button
                      type="button"
                      className="btn btn-secondary btn-small"
                      onClick={() => openAssignModal(form)}
                    >
                      Assign
                    </button>
                    <button
                      type="button"
                      className="btn btn-secondary btn-small"
                      onClick={() => openEditModal(form.id)}
                    >
                      Edit
                    </button>
                    <button
                      type="button"
                      className="btn btn-danger btn-small"
                      onClick={() => handleDelete(form.id)}
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {showAssignModal && assigningForm && (
        <div className="modal-overlay" onClick={closeAssignModal}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Assign "{assigningForm.name}"</h2>
              <button className="close-btn" onClick={closeAssignModal}>&times;</button>
            </div>
            <div className="form-group">
              <label>Batch selector with filter</label>
              <div className="assign-toolbar">
                <input
                  type="text"
                  value={assignSearch}
                  onChange={(e) => setAssignSearch(e.target.value)}
                  placeholder={`Filter ${assigningForm.user_type === 'student' ? 'members' : 'employees'}...`}
                />
                <button type="button" className="btn btn-secondary btn-small" onClick={assignToSelf}>
                  Assign to myself
                </button>
                <button
                  type="button"
                  className="btn btn-secondary btn-small"
                  onClick={() => {
                    const allVisibleSelected = filteredAssigneeOptions.length > 0 && filteredAssigneeOptions.every(row => selectedAssigneeKeys[assigneeKey(row)]);
                    const next = { ...selectedAssigneeKeys };
                    filteredAssigneeOptions.forEach((row) => {
                      const key = assigneeKey(row);
                      next[key] = !allVisibleSelected;
                    });
                    setSelectedAssigneeKeys(next);
                  }}
                >
                  Toggle visible
                </button>
              </div>
            </div>
            <div className="assign-list">
              {assigneeLoading ? (
                <div className="loading">Loading assignees...</div>
              ) : filteredAssigneeOptions.length === 0 ? (
                <div className="empty-state" style={{ padding: '1rem' }}>
                  No assignees match this filter.
                </div>
              ) : (
                filteredAssigneeOptions.map((row) => (
                  <label key={assigneeKey(row)} className="assign-item">
                    <input
                      type="checkbox"
                      checked={!!selectedAssigneeKeys[assigneeKey(row)]}
                      onChange={() => toggleAssignee(row)}
                    />
                    <span>
                      <strong>{row.name || row.user_id}</strong> ({row.user_id})
                      {row.email ? ` · ${row.email}` : ''}
                    </span>
                  </label>
                ))
              )}
            </div>
            <div className="modal-actions">
              <button type="button" className="btn btn-secondary" onClick={closeAssignModal}>
                Cancel
              </button>
              <button type="button" className="btn" onClick={submitAssignments}>
                Assign Selected
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content modal-content--sheet-editor" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{editingForm ? 'Edit Sheet' : 'Create New Sheet'}</h2>
              <button className="close-btn" onClick={closeModal}>&times;</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label htmlFor="formName">Sheet Name *</label>
                <input
                  type="text"
                  id="formName"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label htmlFor="formDescription">Description</label>
                <textarea
                  id="formDescription"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>
              <div className="form-group">
                <label htmlFor="formUserType">User Type *</label>
                <select
                  id="formUserType"
                  value={formData.user_type}
                  onChange={(e) => setFormData({ ...formData, user_type: e.target.value })}
                  required
                >
                  <option value="">Select...</option>
                  <option value="student">Member</option>
                  <option value="staff">Employee</option>
                </select>
              </div>
              <div className="form-group">
                <label>Sheet Fields</label>
                <div className="fields-editor">
                  {formData.fields.map((field, index) => (
                    <div key={index} className="field-item">
                      <div className="field-item-header">
                        <h4>Field {index + 1}</h4>
                        <button
                          type="button"
                          className="btn btn-danger btn-small"
                          onClick={() => removeField(index)}
                        >
                          Remove
                        </button>
                      </div>
                      <div className="field-row">
                        <div className="form-group">
                          <label>Field Name (ID) *</label>
                          <input
                            type="text"
                            value={field.name}
                            onChange={(e) => updateField(index, { name: e.target.value })}
                            placeholder="e.g., name, age"
                            required
                          />
                        </div>
                        <div className="form-group">
                          <label>Label *</label>
                          <input
                            type="text"
                            value={field.label}
                            onChange={(e) => updateField(index, { label: e.target.value })}
                            placeholder="e.g., Full Name"
                            required
                          />
                        </div>
                      </div>
                      <div className="field-row">
                        <div className="form-group">
                          <label>Placeholder</label>
                          <input
                            type="text"
                            value={field.placeholder || ''}
                            onChange={(e) => updateField(index, { placeholder: e.target.value })}
                            placeholder="Placeholder text"
                          />
                        </div>
                        <div className="form-group checkbox-group">
                          <input
                            type="checkbox"
                            checked={field.required}
                            onChange={(e) => updateField(index, { required: e.target.checked })}
                          />
                          <label>Required Field</label>
                        </div>
                      </div>
                    </div>
                  ))}
                  <button
                    type="button"
                    className="btn btn-secondary btn-small"
                    onClick={addField}
                  >
                    + Add Field
                  </button>
                </div>
              </div>
              <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={closeModal}>
                  Cancel
                </button>
                <button type="submit" className="btn">
                  Save Sheet
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default Forms;
