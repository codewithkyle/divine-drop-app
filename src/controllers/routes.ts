import { router, mount } from "@codewithkyle/router";

router.redirect("/", "/decks");
router.add("/decks", "decks-page");
router.add("/edit/{ID}", "edit-deck-page");
router.add("/deck/{ID}", "deck-page");
router.add("/*", "missing-page");

