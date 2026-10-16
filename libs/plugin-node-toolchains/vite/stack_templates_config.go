package vite

// vue is a Vue app: a single-file root component, type-checked by vue-tsc.
var vue = webApp{
	Entry: "main.ts", Typecheck: "vue-tsc -p tsconfig.json --noEmit", Components: []string{"vue"},
	PluginImport: "import vue from '@vitejs/plugin-vue'", PluginCall: "vue()",
	Dependencies:    map[string]string{"vue": "^3.5.44"},
	DevDependencies: map[string]string{"@vitejs/plugin-vue": "^6.0.10", "vue-tsc": "^3.3.13"},
	Files: map[string]string{
		"[[.Dir]]/src/main.ts": `import { createApp } from 'vue'
import { App } from './app/index.ts'

createApp(App).mount('#app')
`,
		"[[.Dir]]/src/app/app.component.vue": `<script setup lang="ts">
import { greet } from '../greeting'

const message = greet('world').message
</script>

<template>
  <h1>{{ message }}</h1>
</template>
`,
		"[[.Dir]]/src/app/index.ts": `export { default as App } from './app.component.vue'
`,
	},
}

// svelte is a Svelte 5 app: a root component, checked by svelte-check.
var svelte = webApp{
	Entry: "main.ts", Typecheck: "svelte-check --tsconfig ./tsconfig.json --fail-on-warnings", Components: []string{"svelte"},
	PluginImport: "import { svelte } from '@sveltejs/vite-plugin-svelte'", PluginCall: "svelte()",
	Dependencies:    map[string]string{"svelte": "^5.57.2"},
	DevDependencies: map[string]string{"@sveltejs/vite-plugin-svelte": "^7.3.1", "svelte-check": "^4.7.7"},
	Files: map[string]string{
		"[[.Dir]]/src/main.ts": `import { mount } from 'svelte'
import { App } from './app/index.ts'

const target = document.getElementById('app')
if (target === null) {
  throw new Error('index.html has no #app')
}
mount(App, { target })
`,
		"[[.Dir]]/src/app/app.component.svelte": `<script lang="ts">
  import { greet } from '../greeting'

  const message = greet('world').message
</script>

<h1>{message}</h1>
`,
		"[[.Dir]]/src/app/index.ts": `export { default as App } from './app.component.svelte'
`,
	},
}

// html is a page with no framework: a handler writes the greeting into it.
var html = webApp{
	Entry: "main.ts", Typecheck: "tsc -p tsconfig.json --noEmit",
	Files: map[string]string{
		"[[.Dir]]/src/main.ts": `import { showGreeting } from './app/index.ts'

const target = document.getElementById('app')
if (target === null) {
  throw new Error('index.html has no #app')
}
showGreeting(target)
`,
		"[[.Dir]]/src/app/show-greeting.handler.ts": `import { greet } from '../greeting'

/**
 * Shows the greeting in the page.
 * @param target - the element to write it into
 */
export function showGreeting (target: HTMLElement): void {
  const heading = document.createElement('h1')
  heading.textContent = greet('world').message
  target.replaceChildren(heading)
}
`,
		"[[.Dir]]/src/app/index.ts": `export { showGreeting } from './show-greeting.handler.ts'
`,
	},
}
