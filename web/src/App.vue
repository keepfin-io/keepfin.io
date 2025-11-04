<template>
  <img alt="Vue logo" src="./assets/logo.png">
  <HelloWorld msg="Welcome to Your Vue.js App"/>
  <InputGroupIcon />
  <div style="margin-top: 2rem;">
    <ShadcnButton>Shadcn Vue Button</ShadcnButton>
  </div>
</template>

<script>
import HelloWorld from './components/HelloWorld.vue'
import InputGroupIcon from './pages/InputGroupIcon.vue'
import ShadcnButton from './components/ui/Button.vue'

export default {
  name: 'App',
  components: {
    HelloWorld,
    InputGroupIcon,
    ShadcnButton
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
