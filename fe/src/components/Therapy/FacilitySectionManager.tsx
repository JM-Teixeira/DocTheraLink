/**
 * FacilitySectionManager Component
 * 
 * Allows facility coordinators to manage facility sections such as:
 * - Electric Section
 * - Pool Section  
 * - Massage Section
 * - Gym Section
 * - etc.
 * 
 * Features:
 * - Add/edit/delete sections
 * - Color coding for visual identification
 * - Section status management
 * - Usage statistics
 */

import React, { useState, useEffect } from 'react';
import { 
  Plus, 
  Edit3, 
  Trash2, 
  MapPin, 
  Palette, 
  Eye, 
  EyeOff,
  Save,
  X
} from 'lucide-react';

interface FacilitySection {
  id: string;
  name: string;
  description: string;
  color: string;
  is_active: boolean;
  created_by: any;
  created_date: string;
}

interface FacilitySectionManagerProps {
  facilityId: string;
  userRole: string;
}

const FacilitySectionManager: React.FC<FacilitySectionManagerProps> = ({ 
  facilityId, 
  userRole 
}) => {
  const [sections, setSections] = useState<FacilitySection[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingSection, setEditingSection] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    color: '#3B82F6',
  });

  // Predefined colors for sections
  const colorOptions = [
    '#3B82F6', // Blue
    '#10B981', // Green
    '#F59E0B', // Yellow
    '#EF4444', // Red
    '#8B5CF6', // Purple
    '#F97316', // Orange
    '#06B6D4', // Cyan
    '#84CC16', // Lime
    '#EC4899', // Pink
    '#6B7280', // Gray
  ];

  useEffect(() => {
    fetchSections();
  }, [facilityId]);

  const fetchSections = async () => {
    setLoading(true);
    try {
      const response = await fetch(`/api/v1/facility/${facilityId}/sections/`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setSections(data.results || []);
      }
    } catch (error) {
      console.error('Failed to fetch sections:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const url = editingSection 
        ? `/api/v1/facility/${facilityId}/sections/${editingSection}/`
        : `/api/v1/facility/${facilityId}/sections/`;
      
      const method = editingSection ? 'PUT' : 'POST';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
        body: JSON.stringify(formData),
      });
      
      if (response.ok) {
        fetchSections();
        resetForm();
      }
    } catch (error) {
      console.error('Failed to save section:', error);
    }
  };

  const handleEdit = (section: FacilitySection) => {
    setEditingSection(section.id);
    setFormData({
      name: section.name,
      description: section.description,
      color: section.color,
    });
    setShowAddForm(true);
  };

  const handleDelete = async (sectionId: string) => {
    if (!confirm('Are you sure you want to delete this section?')) return;
    
    try {
      const response = await fetch(
        `/api/v1/facility/${facilityId}/sections/${sectionId}/`,
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
          },
        }
      );
      
      if (response.ok) {
        fetchSections();
      }
    } catch (error) {
      console.error('Failed to delete section:', error);
    }
  };

  const toggleSectionStatus = async (sectionId: string, isActive: boolean) => {
    try {
      const response = await fetch(
        `/api/v1/facility/${facilityId}/sections/${sectionId}/`,
        {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
          },
          body: JSON.stringify({ is_active: !isActive }),
        }
      );
      
      if (response.ok) {
        fetchSections();
      }
    } catch (error) {
      console.error('Failed to toggle section status:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      color: '#3B82F6',
    });
    setShowAddForm(false);
    setEditingSection(null);
  };

  const canManageSections = () => {
    return ['LocalBodyAdmin', 'DistrictAdmin', 'StateAdmin'].includes(userRole);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      {/* Header */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-gray-900">Facility Sections</h2>
            <p className="text-sm text-gray-500 mt-1">
              Manage therapy sections like Pool, Electric, Massage, etc.
            </p>
          </div>
          
          {canManageSections() && (
            <button
              onClick={() => setShowAddForm(true)}
              className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
            >
              <Plus className="h-4 w-4" />
              <span>Add Section</span>
            </button>
          )}
        </div>
      </div>

      {/* Add/Edit Form */}
      {showAddForm && (
        <div className="p-6 border-b border-gray-200 bg-gray-50">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-gray-900">
                {editingSection ? 'Edit Section' : 'Add New Section'}
              </h3>
              <button
                type="button"
                onClick={resetForm}
                className="p-2 hover:bg-gray-200 rounded-lg transition-colors"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Section Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g., Pool Section, Electric Section"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
                  required
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Color
                </label>
                <div className="flex items-center space-x-2">
                  <div
                    className="w-10 h-10 rounded-lg border-2 border-gray-300"
                    style={{ backgroundColor: formData.color }}
                  />
                  <div className="flex flex-wrap gap-1">
                    {colorOptions.map((color) => (
                      <button
                        key={color}
                        type="button"
                        onClick={() => setFormData({ ...formData, color })}
                        className={`w-6 h-6 rounded border-2 ${
                          formData.color === color ? 'border-gray-800' : 'border-gray-300'
                        }`}
                        style={{ backgroundColor: color }}
                      />
                    ))}
                  </div>
                </div>
              </div>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Description
              </label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Brief description of this section..."
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>
            
            <div className="flex items-center space-x-3">
              <button
                type="submit"
                className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
              >
                <Save className="h-4 w-4" />
                <span>{editingSection ? 'Update' : 'Create'} Section</span>
              </button>
              
              <button
                type="button"
                onClick={resetForm}
                className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Sections List */}
      <div className="p-6">
        {sections.length === 0 ? (
          <div className="text-center py-12">
            <MapPin className="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p className="text-lg font-medium text-gray-500 mb-2">No Sections</p>
            <p className="text-gray-400 mb-4">
              No facility sections have been created yet.
            </p>
            {canManageSections() && (
              <button
                onClick={() => setShowAddForm(true)}
                className="flex items-center space-x-2 mx-auto px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
              >
                <Plus className="h-4 w-4" />
                <span>Add First Section</span>
              </button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {sections.map((section) => (
              <div
                key={section.id}
                className={`p-4 rounded-lg border-2 transition-all duration-200 ${
                  section.is_active 
                    ? 'border-gray-200 bg-white' 
                    : 'border-gray-100 bg-gray-50 opacity-60'
                }`}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center space-x-3">
                    <div
                      className="w-4 h-4 rounded-full border border-gray-300"
                      style={{ backgroundColor: section.color }}
                    />
                    <div>
                      <h3 className="font-medium text-gray-900">{section.name}</h3>
                      <p className="text-sm text-gray-500">
                        Created {new Date(section.created_date).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  
                  {canManageSections() && (
                    <div className="flex items-center space-x-1">
                      <button
                        onClick={() => toggleSectionStatus(section.id, section.is_active)}
                        className="p-1 hover:bg-gray-100 rounded transition-colors"
                        title={section.is_active ? 'Deactivate' : 'Activate'}
                      >
                        {section.is_active ? (
                          <Eye className="h-4 w-4 text-gray-500" />
                        ) : (
                          <EyeOff className="h-4 w-4 text-gray-400" />
                        )}
                      </button>
                      
                      <button
                        onClick={() => handleEdit(section)}
                        className="p-1 hover:bg-gray-100 rounded transition-colors"
                        title="Edit"
                      >
                        <Edit3 className="h-4 w-4 text-gray-500" />
                      </button>
                      
                      <button
                        onClick={() => handleDelete(section.id)}
                        className="p-1 hover:bg-gray-100 rounded transition-colors"
                        title="Delete"
                      >
                        <Trash2 className="h-4 w-4 text-red-500" />
                      </button>
                    </div>
                  )}
                </div>
                
                {section.description && (
                  <p className="text-sm text-gray-600 mb-3">{section.description}</p>
                )}
                
                <div className="flex items-center justify-between text-xs text-gray-500">
                  <span>
                    Status: {section.is_active ? 'Active' : 'Inactive'}
                  </span>
                  <span>
                    By {section.created_by?.first_name} {section.created_by?.last_name}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default FacilitySectionManager;