<template>
  <div class="login-wrapper">
    <div class="login-card">
      <h1>Admin Login</h1>
      <p class="login-subtitle">Sign in to manage guestbook entries</p>

      <div v-if="error" class="alert alert-error">{{ error }}</div>

      <div class="form-group">
        <label>Username</label>
        <input v-model="username" type="text" placeholder="admin" @keyup.enter="login" />
      </div>
      <div class="form-group">
        <label>Password</label>
        <input v-model="password" type="password" placeholder="••••••" @keyup.enter="login" />
      </div>
      <button class="btn btn-primary btn-full" @click="login" :disabled="loading">
        {{ loading ? 'Signing in...' : 'Sign In' }}
      </button>
    </div>
  </div>
</template>

<script>
export default {
  emits: ['login-success'],
  data() {
    return {
      username: '',
      password: '',
      error: '',
      loading: false
    }
  },
  methods: {
    async login() {
      this.error = ''
      if (!this.username || !this.password) {
        this.error = 'Please enter both username and password'
        return
      }

      this.loading = true
      try {
        const res = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            username: this.username,
            password: this.password
          })
        })

        const data = await res.json()

        if (!res.ok) {
          this.error = data.error || 'Login failed'
          return
        }

        this.$emit('login-success', {
          token: data.token,
          username: data.username
        })
      } catch {
        this.error = 'Network error. Is the auth service running?'
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.login-wrapper {
  display: flex;
  justify-content: center;
  padding-top: 3rem;
}

.login-card {
  background: #fff;
  border-radius: 10px;
  padding: 2rem;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
}

.login-card h1 {
  font-size: 1.5rem;
  margin-bottom: 0.25rem;
}

.login-subtitle {
  color: #888;
  font-size: 0.85rem;
  margin-bottom: 1.5rem;
}

.form-group {
  margin-bottom: 1rem;
}
.form-group label {
  display: block;
  font-size: 0.8rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
  color: #444;
}
.form-group input {
  width: 100%;
  padding: 0.6rem 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 0.9rem;
}
.form-group input:focus {
  outline: none;
  border-color: #4361ee;
  box-shadow: 0 0 0 2px rgba(67,97,238,0.15);
}

.btn-full {
  width: 100%;
  margin-top: 0.5rem;
}

.alert {
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}
.alert-error { background: #fde8e8; color: #c53030; }
</style>
