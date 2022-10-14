import { router, mount } from "@codewithkyle/router";

router.redirect("/", "/decks");
router.add("/decks", "decks-page");
router.add("/deck/{ID}", "edit-deck-page");
router.add("/*", "missing-page");

