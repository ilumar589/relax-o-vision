import 'htmx.org'

import 'flowbite'

import Alpine from 'alpinejs'

// Add Alpine instance to window object.
window.Alpine = Alpine

// Start Alpine.
Alpine.start()

// Document ready function to ensure the DOM is fully loaded.
document.addEventListener('DOMContentLoaded', function () {
  initFlowbite() // initialize Flowbite
})

// Add event listeners for all HTMX events.
document.body.addEventListener(
  'htmx:afterSwap htmx:afterRequest htmx:afterSettle',
  function () {
    initFlowbite() // initialize Flowbite
  }
)
