import { ref, computed } from "vue";
import { defineStore } from "pinia";

export const useLogStore = defineStore("log_history", () => {
    const history = ref("");

    function add_log_line(line: string) {
        history.value = line + history.value;

        if (history.value.length >= 10_000) {
            history.value = history.value.slice(0, 10_000);
        }
    }

    return { history, add_log_line };
});
