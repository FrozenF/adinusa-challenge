<template>
  <div id="app">
    <nav class="navbar">
      <div class="nav-container">
        <router-link to="/" class="nav-brand">📖 GuestBook</router-link>
        <div class="nav-links">
          <router-link to="/" class="nav-link">Home</router-link>
          <template v-if="isLoggedIn">
            <span class="nav-user">👤 {{ username }}</span>
            <button class="btn btn-sm btn-outline" @click="logout">Logout</button>
          </template>
          <router-link v-else to="/login" class="btn btn-sm btn-primary">Login</router-link>
        </div>
      </div>
    </nav>
    <main class="container">
      <router-view
        :is-logged-in="isLoggedIn"
        :token="token"
        @login-success="onLoginSuccess"
      />
    </main>
  </div>
</template>

<script>
export default {
  data() {
    return {
      token: localStorage.getItem('session_token') || '',
      username: localStorage.getItem('username') || ''
    }
  },
  computed: {
    isLoggedIn() {
      return !!this.token
    }
  },
  async mounted() {
    if (this.token) {
      try {
        const res = await fetch('/api/auth/me', {
          headers: { 'Authorization': `Bearer ${this.token}` }
        })
        if (!res.ok) {
          this.clearSession()
        }
      } catch {
        this.clearSession()
      }
    }
  },
  methods: {
    onLoginSuccess({ token, username }) {
      this.token = token
      this.username = username
      localStorage.setItem('session_token', token)
      localStorage.setItem('username', username)
      this.$router.push('/')
    },
    async logout() {
      try {
        await fetch('/api/auth/logout', {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${this.token}` }
        })
      } catch {
        // ignore
      }
      this.clearSession()
      this.$router.push('/')
    },
    clearSession() {
      this.token = ''
      this.username = ''
      localStorage.removeItem('session_token')
      localStorage.removeItem('username')
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f5f7fa;
  color: #1a1a2e;
  line-height: 1.6;
}

.navbar {
  background: #1a1a2e;
  color: #fff;
  padding: 0 1rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
}

.nav-container {
  max-width: 900px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
}

.nav-brand {
  font-size: 1.25rem;
  font-weight: 700;
  color: #fff;
  text-decoration: none;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-link {
  color: #ccc;
  text-decoration: none;
  font-size: 0.9rem;
}
.nav-link:hover { color: #fff; }

.nav-user {
  font-size: 0.85rem;
  color: #a0d2db;
}

.container {
  max-width: 900px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.btn {
  padding: 0.5rem 1.25rem;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
  font-weight: 500;
  transition: background 0.2s;
}
.btn-sm { padding: 0.35rem 0.9rem; font-size: 0.8rem; }
.btn-primary { background: #4361ee; color: #fff; }
.btn-primary:hover { background: #3a56d4; }
.btn-danger { background: #e63946; color: #fff; }
.btn-danger:hover { background: #c5303c; }
.btn-outline {
  background: transparent;
  color: #fff;
  border: 1px solid #fff;
}
.btn-outline:hover { background: rgba(255,255,255,0.1); }
</style>
