<script lang="ts">
  import Router from "svelte-spa-router";
  import Navbar from "./lib/Navbar.svelte";
  import Login from "./lib/Login.svelte";
  import Register from "./lib/Register.svelte";
  import Dashboard from "./lib/Dashboard.svelte";

  const routes = {
    '/': Dashboard,
    '/login': Login,
    '/register': Register,
    // Catch-all
    '*': Dashboard
  };
</script>

<main>
  <Navbar />
  <Router {routes} />
</main>