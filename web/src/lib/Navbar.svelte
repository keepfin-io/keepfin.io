<script lang="ts">
  import { link } from "svelte-spa-router";
</script>

<nav style="display: flex; gap: 1rem; padding: 1rem; background-color: #333;">
  <a href="/" use:link style="color: white; text-decoration: none;">Dashboard</a>
  <a href="/login" use:link style="color: white; text-decoration: none;">Login</a>
  <a href="/register" use:link style="color: white; text-decoration: none;">Register</a>
</nav>