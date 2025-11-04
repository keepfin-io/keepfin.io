<template>
  <div class="grid w-full max-w-sm gap-6">
    <InputGroup>
      <template #input>
        <InputGroupInput placeholder="Search..." />
      </template>
      <template #addon>
        <InputGroupAddon>
          <LucideSearch />
        </InputGroupAddon>
      </template>
    </InputGroup>
    <InputGroup>
      <template #input>
        <InputGroupInput type="email" placeholder="Enter your email" />
      </template>
      <template #addon>
        <InputGroupAddon>
          <LucideMail />
        </InputGroupAddon>
      </template>
    </InputGroup>
    <InputGroup>
      <template #input>
        <InputGroupInput placeholder="Card number" />
      </template>
      <template #addon>
        <InputGroupAddon>
          <LucideCreditCard />
        </InputGroupAddon>
        <InputGroupAddon>
          <LucideCheck />
        </InputGroupAddon>
      </template>
    </InputGroup>
    <InputGroup>
      <template #input>
        <InputGroupInput placeholder="Card number" />
      </template>
      <template #addon>
        <InputGroupAddon>
          <LucideStar />
          <LucideInfo />
        </InputGroupAddon>
      </template>
    </InputGroup>
  </div>
</template>

<script setup>
import InputGroup from '../components/InputGroup.vue'
import InputGroupInput from '../components/InputGroupInput.vue'
import InputGroupAddon from '../components/InputGroupAddon.vue'
import { LucideSearch, LucideMail, LucideCreditCard, LucideCheck, LucideStar, LucideInfo } from 'lucide-vue-next'
</script>
