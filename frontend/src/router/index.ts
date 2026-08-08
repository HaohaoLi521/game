import { createRouter, createWebHistory } from "vue-router";
import HomeView from "../views/HomeView.vue";
import GameView from "../views/GameView.vue";
import PuzzleSelectView from "../views/PuzzleSelectView.vue";
import ResultView from "../views/ResultView.vue";
import SettingsView from "../views/SettingsView.vue";
import AccountCenterView from "../views/AccountCenterView.vue";
import SubmitView from "../views/SubmitView.vue";
import MySubmissionsView from "../views/MySubmissionsView.vue";
import WorkshopView from "../views/WorkshopView.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: HomeView },
    { path: "/settings", name: "settings", component: SettingsView },
    { path: "/account", name: "account", component: AccountCenterView },
    { path: "/puzzles", name: "puzzles", component: PuzzleSelectView },
    { path: "/game", name: "game", component: GameView },
    { path: "/result", name: "result", component: ResultView },
    { path: "/submit", name: "submit", component: SubmitView },
    { path: "/my-submissions", name: "my-submissions", component: MySubmissionsView },
    { path: "/workshop", name: "workshop", component: WorkshopView }
  ]
});

export default router;
