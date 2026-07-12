import { createRouter, createWebHistory } from "vue-router";
import HomeView from "../views/HomeView.vue";
import GameView from "../views/GameView.vue";
import PuzzleSelectView from "../views/PuzzleSelectView.vue";
import ResultView from "../views/ResultView.vue";
import SubmitView from "../views/SubmitView.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: HomeView },
    { path: "/puzzles", name: "puzzles", component: PuzzleSelectView },
    { path: "/game", name: "game", component: GameView },
    { path: "/result", name: "result", component: ResultView },
    { path: "/submit", name: "submit", component: SubmitView }
  ]
});

export default router;