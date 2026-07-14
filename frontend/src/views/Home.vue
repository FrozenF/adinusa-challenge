<template>
  <div>
    <div class="page-header">
      <h1>Guest Book</h1>
      <p class="subtitle">Leave a message for us!</p>
    </div>

    <!-- Add Entry Form (only when logged in) -->
    <div v-if="isLoggedIn" class="card form-card">
      <h2>Add New Entry</h2>
      <div v-if="formError" class="alert alert-error">{{ formError }}</div>
      <div v-if="formSuccess" class="alert alert-success">{{ formSuccess }}</div>
      <div class="form-grid">
        <div class="form-group">
          <label>Name</label>
          <input v-model="form.name" type="text" placeholder="Your name" />
        </div>
        <div class="form-group">
          <label>Address</label>
          <input v-model="form.address" type="text" placeholder="Your city / address" />
        </div>
        <div class="form-group full-width">
          <label>Message</label>
          <textarea v-model="form.message" placeholder="Write your message..." rows="3"></textarea>
        </div>
      </div>
      <button class="btn btn-primary" @click="addEntry" :disabled="submitting">
        {{ submitting ? 'Submitting...' : 'Submit Entry' }}
      </button>
    </div>

    <!-- Entries List -->
    <div v-if="loading" class="loading">Loading entries...</div>
    <div v-else-if="entries.length === 0" class="empty-state">
      <p>No guestbook entries yet. Be the first!</p>
    </div>
    <div v-else class="entries-list">
      <div v-for="entry in entries" :key="entry.id" class="card entry-card">
        <div class="entry-header">
          <div>
            <strong class="entry-name">{{ entry.name }}</strong>
            <span class="entry-address">📍 {{ entry.address }}</span>
          </div>
          <div class="entry-meta">
            <span class="entry-date">{{ formatDate(entry.created_at) }}</span>
            <button
              v-if="isLoggedIn"
              class="btn btn-danger btn-sm"
              @click="deleteEntry(entry.id)"
            >Delete</button>
          </div>
        </div>
        <p class="entry-message">{{ entry.message }}</p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    isLoggedIn: Boolean,
    token: String
  },
  data() {
    return {
      entries: [],
      loading: true,
      submitting: false,
      formError: '',
      formSuccess: '',
      form: { name: '', address: '', message: '' }
    }
  },
  async mounted() {
    await this.fetchEntries()
  },
  methods: {
    async fetchEntries() {
      this.loading = true
      try {
        const res = await fetch('/api/guestbook')
        this.entries = await res.json()
      } catch {
        this.entries = []
      }
      this.loading = false
    },
    async addEntry() {
      this.formError = ''
      this.formSuccess = ''

      if (!this.form.name || !this.form.address || !this.form.message) {
        this.formError = 'All fields are required'
        return
      }

      this.submitting = true
      try {
        const res = await fetch('/api/guestbook', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
          },
          body: JSON.stringify(this.form)
        })

        if (!res.ok) {
          const data = await res.json()
          this.formError = data.error || 'Failed to add entry'
          return
        }

        this.formSuccess = 'Entry added successfully!'
        this.form = { name: '', address: '', message: '' }
        await this.fetchEntries()

        setTimeout(() => { this.formSuccess = '' }, 3000)
      } catch {
        this.formError = 'Network error. Please try again.'
      } finally {
        this.submitting = false
      }
    },
    async deleteEntry(id) {
      if (!confirm('Are you sure you want to delete this entry?')) return
      try {
        await fetch(`/api/guestbook/${id}`, {
          method: 'DELETE',
          headers: { 'Authorization': `Bearer ${this.token}` }
        })
        await this.fetchEntries()
      } catch {
        alert('Failed to delete entry')
      }
    },
    formatDate(dateStr) {
      return new Date(dateStr).toLocaleDateString('en-US', {
        year: 'numeric', month: 'short', day: 'numeric',
        hour: '2-digit', minute: '2-digit'
      })
    }
  }
}
</script>

<style scoped>
.page-header {
  margin-bottom: 1.5rem;
}
.page-header h1 {
  font-size: 1.75rem;
  color: #1a1a2e;
}
.subtitle {
  color: #666;
  font-size: 0.95rem;
}

.card {
  background: #fff;
  border-radius: 10px;
  padding: 1.25rem;
  margin-bottom: 1rem;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

.form-card h2 {
  font-size: 1.1rem;
  margin-bottom: 1rem;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  margin-bottom: 1rem;
}
.full-width { grid-column: 1 / -1; }

.form-group label {
  display: block;
  font-size: 0.8rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
  color: #444;
}
.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 0.9rem;
  font-family: inherit;
}
.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #4361ee;
  box-shadow: 0 0 0 2px rgba(67,97,238,0.15);
}

.alert {
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}
.alert-error { background: #fde8e8; color: #c53030; }
.alert-success { background: #e6ffed; color: #22543d; }

.loading, .empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #888;
}

.entry-card { transition: box-shadow 0.2s; }
.entry-card:hover { box-shadow: 0 2px 10px rgba(0,0,0,0.1); }

.entry-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.entry-name {
  font-size: 1rem;
  color: #1a1a2e;
  margin-right: 0.5rem;
}

.entry-address {
  font-size: 0.8rem;
  color: #888;
}

.entry-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.entry-date {
  font-size: 0.75rem;
  color: #aaa;
}

.entry-message {
  color: #444;
  font-size: 0.9rem;
  white-space: pre-wrap;
}
</style>
