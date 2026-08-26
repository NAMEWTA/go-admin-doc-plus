import { mapState } from 'pinia'
import { useAppStore } from '@/stores/app'

export default {
  computed: {
    ...mapState(useAppStore, ['device'])
  },
  mounted() {
    // Keep an iOS menu tap from being treated as a desktop mouseleave.
    this.fixBugIniOS()
  },
  methods: {
    fixBugIniOS() {
      const $subMenu = this.$refs.subMenu
      if ($subMenu && typeof $subMenu.handleMouseleave === 'function') {
        const handleMouseleave = $subMenu.handleMouseleave
        $subMenu.handleMouseleave = (e) => {
          if (this.device === 'mobile') {
            return
          }
          handleMouseleave(e)
        }
      }
    }
  }
}
