import { defineStore } from "pinia";
import type { AnswerMode } from "../api/puzzles";

export type AnswerModePreference = "auto" | AnswerMode;

interface SettingsState {
  answerMode: AnswerModePreference;
  soundEnabled: boolean;
  vibrationEnabled: boolean;
}

const settingsKey = "this-is-pun-settings";

function readSettings(): SettingsState {
  const defaults: SettingsState = { answerMode: "auto", soundEnabled: true, vibrationEnabled: true };
  try {
    const value = JSON.parse(localStorage.getItem(settingsKey) || "{}");
    return {
      answerMode: ["auto", "manual", "tiles"].includes(value.answerMode) ? value.answerMode : defaults.answerMode,
      soundEnabled: typeof value.soundEnabled === "boolean" ? value.soundEnabled : defaults.soundEnabled,
      vibrationEnabled: typeof value.vibrationEnabled === "boolean" ? value.vibrationEnabled : defaults.vibrationEnabled
    };
  } catch {
    return defaults;
  }
}

function saveSettings(settings: SettingsState) {
  localStorage.setItem(settingsKey, JSON.stringify(settings));
}

export const useSettingsStore = defineStore("settings", {
  state: (): SettingsState => readSettings(),

  actions: {
    setAnswerMode(answerMode: AnswerModePreference) {
      this.answerMode = answerMode;
      saveSettings(this.$state);
    },

    setSoundEnabled(soundEnabled: boolean) {
      this.soundEnabled = soundEnabled;
      saveSettings(this.$state);
    },

    setVibrationEnabled(vibrationEnabled: boolean) {
      this.vibrationEnabled = vibrationEnabled;
      saveSettings(this.$state);
    }
  }
});
