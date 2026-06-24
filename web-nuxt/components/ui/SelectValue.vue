<script setup lang="ts">
import { SelectValue, type SelectValueProps } from 'radix-vue'
const props = defineProps<SelectValueProps>()
</script>

<SelectValue v-bind="props">
  <slot />
</SelectValue>