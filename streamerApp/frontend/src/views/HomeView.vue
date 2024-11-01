<script setup lang="ts">
import { ref } from 'vue';
import { onMounted } from 'vue';
import { Greet, GetUserConfig, Connect, Disconnect } from "../../wailsjs/go/main/App";
import { useStatusStore } from '@/stores/status';

const name = ref("hi")
const response = ref("response...")
const user_config = ref({
  TwitchUsername: "",
  ObsHost: "",
  ObsPort: "",
  ObsPassword: ""
})

function getGreet() {
  Greet(name.value).then((res) => response.value = res);
}
function DoConnect() {
  Connect().then(() => { });
}
function DoDisconnect() {
  Disconnect().then(() => { });
}

const status = useStatusStore();

onMounted(() => {
  GetUserConfig().then((data) => user_config.value = data)
});
</script>

<template>
  <section id="config-form">
    <form>
      <div class="form-row">
        <label for="twitch-username">Twitch Username</label>
        <input type="text" id="twitch-username" v-model="user_config.TwitchUsername"
          placeholder="Enter your Twitch username" />
      </div>
      <div class="form-row">
        <label for="obs-host">OBS Host</label>
        <input type="text" id="obs-host" v-model="user_config.ObsHost" placeholder="e.g., localhost" />
      </div>
      <div class="form-row">
        <label for="obs-port">OBS Port</label>
        <input type="number" id="obs-port" v-model="user_config.ObsPort" placeholder="e.g., 4444" />
      </div>

      <div class="form-row">
        <label for="obs-password">OBS Password</label>
        <input type="password" id="obs-password" v-model="user_config.ObsPassword"
          placeholder="Enter your OBS password" />
      </div>
      <!-- <div class="form-row"> -->
      <div class="button-container">
        <button type="submit" class="save-button">Save</button>
      </div>
      <!-- </div> -->

    </form>
    <div style="width:100%; margin: auto">
      <form>
        <button v-if="status.status == 'Disconnected'"
          style="margin-top: 16px; padding:16px; width:100%; color: white; background-color: #3e8e41;" type="button"
          class="save-button" @click="DoConnect">Connect</button>
        <button v-else style="margin-top: 16px; padding:16px; width:100%; color: white; background-color: #dd3333;"
          type="button" class="save-button" @click="DoDisconnect">Disconnect</button>
      </form>
    </div>
  </section>

</template>

<style lang="scss">
// section {
//   padding: 32px;
//   color:#ffffff;

//   input {
//     padding: 4px;
//     color: #000000;
//   }
//   .form_rows {
//     display: flex;
//     flex-flow: column;
//   }
//   .row_item {
//     padding: 8px;
//     label {
//       padding: 28px;
//     }
//   }

//   button {
//     display: inline-block;
//     min-width: 80px;
//     height: 30px;
//     line-height: 30px;
//     padding: 0 5px;
//     margin-left: 8px;
//     background-color: #3459b0;
//     border-radius: 2px;
//     text-align: center;
//     text-decoration: none;
//     color: #000000;
//     font-size: 14px;

//     &:hover {
//       background-color: #4879eb;
//       color: #ffffff;
//       cursor: pointer;
//     }
//   }
// }
#config-form {
  max-width: 500px;
  margin: 20px auto;
  padding: 16px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
}

.form-row {
  display: flex;
  margin-bottom: 12px;
  align-items: center;
}

.form-row label {
  flex: 1;
  font-weight: bold;
  margin-right: 10px;
  min-width: 120px;
}

.form-row input {
  flex: 2;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}

.form-row input[type="password"] {
  font-family: sans-serif;
  letter-spacing: 1px;
}

.button-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.button-container .save-button {
  padding: 10px 20px;
  background-color: #4CAF50;
  color: rgb(255, 255, 255);
  border: none;
  border-radius: 4px;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  transition: background-color 0.3s;
}

.save-button:hover {
  background-color: #45a049;
}

.save-button:active {
  background-color: #3e8e41;
}
</style>
