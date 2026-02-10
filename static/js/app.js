// Email Reminder System - Client-side JavaScript

// Form validation helpers
const FormValidator = {
    // Validate email address format
    isValidEmail: function(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email.trim());
    },

    // Validate time format (HH:MM)
    isValidTime: function(time) {
        const timeRegex = /^([01][0-9]|2[0-3]):[0-5][0-9]$/;
        return timeRegex.test(time.trim());
    },

    // Validate required field
    isRequired: function(value) {
        return value && value.trim().length > 0;
    },

    // Validate number range
    isInRange: function(value, min, max) {
        const num = parseInt(value, 10);
        return !isNaN(num) && num >= min && num <= max;
    },

    // Show error message on field
    showError: function(field, message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'field-error';
        errorDiv.textContent = message;
        errorDiv.style.color = '#c33';
        errorDiv.style.fontSize = '12px';
        errorDiv.style.marginTop = '5px';
        
        // Remove existing error if present
        this.clearError(field);
        
        // Add error message
        field.parentElement.appendChild(errorDiv);
        field.style.borderColor = '#c33';
    },

    // Clear error message from field
    clearError: function(field) {
        const existingError = field.parentElement.querySelector('.field-error');
        if (existingError) {
            existingError.remove();
        }
        field.style.borderColor = '';
    }
};

// Loading state management
const LoadingState = {
    // Show loading state on button
    showLoading: function(button) {
        button.disabled = true;
        button.dataset.originalText = button.textContent;
        button.textContent = 'İşleniyor...';
        button.classList.add('loading');
    },

    // Hide loading state on button
    hideLoading: function(button) {
        button.disabled = false;
        if (button.dataset.originalText) {
            button.textContent = button.dataset.originalText;
            delete button.dataset.originalText;
        }
        button.classList.remove('loading');
    },

    // Show loading overlay on form
    showFormLoading: function(form) {
        form.classList.add('loading');
        const buttons = form.querySelectorAll('button[type="submit"]');
        buttons.forEach(btn => this.showLoading(btn));
    },

    // Hide loading overlay on form
    hideFormLoading: function(form) {
        form.classList.remove('loading');
        const buttons = form.querySelectorAll('button[type="submit"]');
        buttons.forEach(btn => this.hideLoading(btn));
    }
};

// Confirmation dialogs
const ConfirmDialog = {
    // Show confirmation dialog for delete operations
    confirmDelete: function(message) {
        return confirm(message || 'Bu işlemi gerçekleştirmek istediğinizden emin misiniz?');
    }
};

// Initialize form validation on page load
document.addEventListener('DOMContentLoaded', function() {
    // Validate reminder form if present
    const reminderForm = document.querySelector('form[action*="/reminders"]');
    if (reminderForm && !reminderForm.querySelector('button[type="submit"]').textContent.includes('Sil')) {
        initReminderFormValidation(reminderForm);
    }

    // Validate settings form if present
    const settingsForm = document.querySelector('form[action="/settings"]');
    if (settingsForm) {
        initSettingsFormValidation(settingsForm);
    }

    // Add confirmation to delete buttons
    const deleteForms = document.querySelectorAll('form[action*="/delete"]');
    deleteForms.forEach(form => {
        form.addEventListener('submit', function(e) {
            if (!ConfirmDialog.confirmDelete('Bu hatırlatıcıyı silmek istediğinizden emin misiniz?')) {
                e.preventDefault();
            }
        });
    });

    // Add loading state to all forms
    const allForms = document.querySelectorAll('form');
    allForms.forEach(form => {
        form.addEventListener('submit', function() {
            // Don't show loading for delete confirmations that were cancelled
            if (form.action.includes('/delete')) {
                return;
            }
            LoadingState.showFormLoading(form);
        });
    });
});

// Initialize reminder form validation
function initReminderFormValidation(form) {
    const titleField = form.querySelector('#title');
    const recipientsField = form.querySelector('#recipients');
    const emailContentField = form.querySelector('#email_content');
    const scheduleTypeField = form.querySelector('#schedule_type');
    const timeOfDayField = form.querySelector('#time_of_day');
    const dayOfWeekField = form.querySelector('#day_of_week');
    const dayOfMonthField = form.querySelector('#day_of_month');
    const intervalDaysField = form.querySelector('#interval_days');

    // Real-time validation for title
    if (titleField) {
        titleField.addEventListener('blur', function() {
            if (!FormValidator.isRequired(this.value)) {
                FormValidator.showError(this, 'Başlık gereklidir');
            } else {
                FormValidator.clearError(this);
            }
        });
    }

    // Real-time validation for recipients
    if (recipientsField) {
        recipientsField.addEventListener('blur', function() {
            const emails = this.value.split(',').map(e => e.trim()).filter(e => e.length > 0);
            if (emails.length === 0) {
                FormValidator.showError(this, 'En az bir alıcı e-posta adresi gereklidir');
            } else {
                const invalidEmails = emails.filter(email => !FormValidator.isValidEmail(email));
                if (invalidEmails.length > 0) {
                    FormValidator.showError(this, `Geçersiz e-posta adresleri: ${invalidEmails.join(', ')}`);
                } else {
                    FormValidator.clearError(this);
                }
            }
        });
    }

    // Real-time validation for email content
    if (emailContentField) {
        emailContentField.addEventListener('blur', function() {
            if (!FormValidator.isRequired(this.value)) {
                FormValidator.showError(this, 'E-posta içeriği gereklidir');
            } else {
                FormValidator.clearError(this);
            }
        });
    }

    // Real-time validation for time of day
    if (timeOfDayField) {
        timeOfDayField.addEventListener('blur', function() {
            if (!FormValidator.isRequired(this.value)) {
                FormValidator.showError(this, 'Gönderim saati gereklidir');
            } else if (!FormValidator.isValidTime(this.value)) {
                FormValidator.showError(this, 'Geçerli bir saat formatı girin (HH:MM, örn: 14:30)');
            } else {
                FormValidator.clearError(this);
            }
        });
    }

    // Real-time validation for day of month
    if (dayOfMonthField) {
        dayOfMonthField.addEventListener('blur', function() {
            if (scheduleTypeField.value === 'monthly') {
                if (!FormValidator.isRequired(this.value)) {
                    FormValidator.showError(this, 'Ayın günü gereklidir');
                } else if (!FormValidator.isInRange(this.value, 1, 31)) {
                    FormValidator.showError(this, 'Ayın günü 1-31 arasında olmalıdır');
                } else {
                    FormValidator.clearError(this);
                }
            }
        });
    }

    // Real-time validation for interval days
    if (intervalDaysField) {
        intervalDaysField.addEventListener('blur', function() {
            if (scheduleTypeField.value === 'custom') {
                if (!FormValidator.isRequired(this.value)) {
                    FormValidator.showError(this, 'Gün aralığı gereklidir');
                } else if (!FormValidator.isInRange(this.value, 1, 365)) {
                    FormValidator.showError(this, 'Gün aralığı 1-365 arasında olmalıdır');
                } else {
                    FormValidator.clearError(this);
                }
            }
        });
    }

    // Form submission validation
    form.addEventListener('submit', function(e) {
        let isValid = true;

        // Validate title
        if (titleField && !FormValidator.isRequired(titleField.value)) {
            FormValidator.showError(titleField, 'Başlık gereklidir');
            isValid = false;
        }

        // Validate recipients
        if (recipientsField) {
            const emails = recipientsField.value.split(',').map(e => e.trim()).filter(e => e.length > 0);
            if (emails.length === 0) {
                FormValidator.showError(recipientsField, 'En az bir alıcı e-posta adresi gereklidir');
                isValid = false;
            } else {
                const invalidEmails = emails.filter(email => !FormValidator.isValidEmail(email));
                if (invalidEmails.length > 0) {
                    FormValidator.showError(recipientsField, `Geçersiz e-posta adresleri: ${invalidEmails.join(', ')}`);
                    isValid = false;
                }
            }
        }

        // Validate email content
        if (emailContentField && !FormValidator.isRequired(emailContentField.value)) {
            FormValidator.showError(emailContentField, 'E-posta içeriği gereklidir');
            isValid = false;
        }

        // Validate schedule type
        if (scheduleTypeField && !FormValidator.isRequired(scheduleTypeField.value)) {
            FormValidator.showError(scheduleTypeField, 'Zamanlama türü seçiniz');
            isValid = false;
        }

        // Validate time of day
        if (timeOfDayField) {
            if (!FormValidator.isRequired(timeOfDayField.value)) {
                FormValidator.showError(timeOfDayField, 'Gönderim saati gereklidir');
                isValid = false;
            } else if (!FormValidator.isValidTime(timeOfDayField.value)) {
                FormValidator.showError(timeOfDayField, 'Geçerli bir saat formatı girin (HH:MM, örn: 14:30)');
                isValid = false;
            }
        }

        // Validate schedule-specific fields
        if (scheduleTypeField) {
            const scheduleType = scheduleTypeField.value;
            
            if (scheduleType === 'monthly' && dayOfMonthField) {
                if (!FormValidator.isRequired(dayOfMonthField.value)) {
                    FormValidator.showError(dayOfMonthField, 'Ayın günü gereklidir');
                    isValid = false;
                } else if (!FormValidator.isInRange(dayOfMonthField.value, 1, 31)) {
                    FormValidator.showError(dayOfMonthField, 'Ayın günü 1-31 arasında olmalıdır');
                    isValid = false;
                }
            }

            if (scheduleType === 'custom' && intervalDaysField) {
                if (!FormValidator.isRequired(intervalDaysField.value)) {
                    FormValidator.showError(intervalDaysField, 'Gün aralığı gereklidir');
                    isValid = false;
                } else if (!FormValidator.isInRange(intervalDaysField.value, 1, 365)) {
                    FormValidator.showError(intervalDaysField, 'Gün aralığı 1-365 arasında olmalıdır');
                    isValid = false;
                }
            }
        }

        if (!isValid) {
            e.preventDefault();
            LoadingState.hideFormLoading(form);
        }
    });
}

// Initialize settings form validation
function initSettingsFormValidation(form) {
    const hostField = form.querySelector('#host');
    const portField = form.querySelector('#port');
    const usernameField = form.querySelector('#username');
    const fromEmailField = form.querySelector('#from_email');

    // Real-time validation for host
    if (hostField) {
        hostField.addEventListener('blur', function() {
            if (!FormValidator.isRequired(this.value)) {
                FormValidator.showError(this, 'SMTP sunucusu gereklidir');
            } else {
                FormValidator.clearError(this);
            }
        });
    }

    // Real-time validation for port
    if (portField) {
        portField.addEventListener('blur', function() {
            if (!FormValidator.isRequired(this.value)) {
                FormValidator.showError(this, 'Port gereklidir');
            } else if (!FormValidator.isInRange(this.value, 1, 65535)) {
                FormValidator.showError(this, 'Port 1-65535 arasında olmalıdır');
            } else {
                FormValidator.clearError(this);
            }
        });
    }

    // Real-time validation for username
    if (usernameField) {
        usernameField.addEventListener('blur', function() {
            if (!FormValidator.isRequired(this.value)) {
                FormValidator.showError(this, 'Kullanıcı adı gereklidir');
            } else {
                FormValidator.clearError(this);
            }
        });
    }

    // Real-time validation for from email
    if (fromEmailField) {
        fromEmailField.addEventListener('blur', function() {
            if (!FormValidator.isRequired(this.value)) {
                FormValidator.showError(this, 'Gönderen e-posta gereklidir');
            } else if (!FormValidator.isValidEmail(this.value)) {
                FormValidator.showError(this, 'Geçerli bir e-posta adresi girin');
            } else {
                FormValidator.clearError(this);
            }
        });
    }

    // Form submission validation
    form.addEventListener('submit', function(e) {
        let isValid = true;

        // Validate host
        if (hostField && !FormValidator.isRequired(hostField.value)) {
            FormValidator.showError(hostField, 'SMTP sunucusu gereklidir');
            isValid = false;
        }

        // Validate port
        if (portField) {
            if (!FormValidator.isRequired(portField.value)) {
                FormValidator.showError(portField, 'Port gereklidir');
                isValid = false;
            } else if (!FormValidator.isInRange(portField.value, 1, 65535)) {
                FormValidator.showError(portField, 'Port 1-65535 arasında olmalıdır');
                isValid = false;
            }
        }

        // Validate username
        if (usernameField && !FormValidator.isRequired(usernameField.value)) {
            FormValidator.showError(usernameField, 'Kullanıcı adı gereklidir');
            isValid = false;
        }

        // Validate from email
        if (fromEmailField) {
            if (!FormValidator.isRequired(fromEmailField.value)) {
                FormValidator.showError(fromEmailField, 'Gönderen e-posta gereklidir');
                isValid = false;
            } else if (!FormValidator.isValidEmail(fromEmailField.value)) {
                FormValidator.showError(fromEmailField, 'Geçerli bir e-posta adresi girin');
                isValid = false;
            }
        }

        if (!isValid) {
            e.preventDefault();
            LoadingState.hideFormLoading(form);
        }
    });
}
