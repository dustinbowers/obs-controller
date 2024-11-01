import { ref } from "vue";
import { defineStore } from "pinia";

export const useStatusStore = defineStore("connection_status", () => {
    const status = ref("");

    function update_status(new_status: string) {
        status.value = new_status
    }

    return { status, update_status };
});
