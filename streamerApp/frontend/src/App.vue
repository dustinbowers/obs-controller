<script setup lang="ts">

import { EventsOn } from "../wailsjs/runtime/runtime"
import { useLogStore } from "@/stores/logs"
import { useStatusStore } from "@/stores/status"

document.body.addEventListener("click", function (event) {
  event.preventDefault();
});

const log_history = useLogStore();
const status = useStatusStore();

EventsOn("log_event", (msg: string) => log_history.add_log_line(msg));
EventsOn("connection_status", (msg: string) => status.update_status(msg))

</script>

<template>
  <!-- Header -->
  <div class="header">
    <!-- navigation -->
    <div class="nav">
      <router-link to="/">Main</router-link>
      <router-link to="/settings">Settings</router-link>
      <router-link to="/logs">Logs</router-link>
    </div>
    <!-- Menu -->
    <div class="menu">
      <div class="bar">
        Status:<div @click=""
          :class="{ 'bar-btn': true, 'connected': status.status != 'Disconnected', 'disconnected': status.status == 'Disconnected' }">
          {{
            status.status }}
        </div>
      </div>
    </div>
  </div>
  <!-- Page -->
  <div class="view">
    <router-view />
  </div>
</template>

<style lang="scss">
@import url("./assets/css/reset.css");
@import url("./assets/css/font.css");

.disconnected {
  background-color: red;
}

.connected {
  background-color: #00862d;
}

html {
  width: 100%;
  height: 100%;
}

body {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
  font-family: "JetBrainsMono";
  // background-color: transparent;
}

#app {
  position: relative;
  // width: 900px;
  // height: 520px;
  height: 100%;
  background-color: rgba(47, 69, 82, 0.9);
  overflow: hidden;
}

.header {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 10px;
  color: #ffffff;
  background-color: rgba(30, 45, 53, 0.9);

  .nav {
    a {
      display: inline-block;
      min-width: 50px;
      height: 42px;
      line-height: 42px;
      padding: 0 64px;
      margin-right: 16px;
      // background-color: #ab7edc;
      background-color: rgba(24, 85, 90, 0.9);
      border-radius: 2px;
      text-align: center;
      text-decoration: none;
      color: #ffffff;
      font-size: 16px;
      white-space: nowrap;

      &:hover,
      &.router-link-exact-active {
        background-color: rgba(43, 148, 155, 0.9);
        border-radius: 2px;
        color: #ffffff;
      }
    }
  }

  .menu {
    display: flex;
    flex-direction: row;
    flex-wrap: nowrap;
    align-items: center;
    justify-content: space-between;

    .bar {
      display: flex;
      flex-direction: row;
      flex-wrap: nowrap;
      align-items: center;
      justify-content: flex-end;
      min-width: 150px;

      .bar-btn {
        display: inline-block;
        min-width: 80px;
        height: 30px;
        line-height: 30px;
        padding: 0 16px;
        margin-left: 8px;
        border-radius: 2px;
        text-align: center;
        text-decoration: none;
        color: #ffffff;
        font-size: 14px;

        &:hover {
          background-color: #4879eb;
          color: #ffffff;
          cursor: pointer;
        }
      }
    }
  }
}

.view {
  position: absolute;
  top: 64px;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
}
</style>
