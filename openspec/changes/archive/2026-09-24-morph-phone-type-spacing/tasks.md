## 1. Failing checks

- [x] 1.1 Add `morph/frontend/src/phoneType.test.js` asserting phone body size, chat bubble and empty-state type, title ellipsis width, 44px error dismiss, login copy size, and notes alert close
- [x] 1.2 Run that test and confirm it fails because the rules and controls are absent

## 2. Type and spacing

- [x] 2.1 Set phone `body` to 16px and line-height 1.5 in `index.css`
- [x] 2.2 In `theme.js`, match that body size after CssBaseline, bump phone `body2`, and give phone alerts, snackbars, and small buttons a 16px / 44px treatment
- [x] 2.3 In the last `App.css` phone block, set chat and skills empty copy to 16px / 1.5, wrap the welcome, and let the header title ellipsize at `width: 100%`

## 3. Empty and error dismiss

- [x] 3.1 Raise sign-in label, helper, and error type to 16px / 1.5, make Sign in at least 44px tall, and add a 44px dismiss control
- [x] 3.2 Add a 44px dismiss control on the chat error bubble without changing the composer
- [x] 3.3 Add `onClose` to the full-page load alerts in `CaseTasks.js` and `StoryBoard.js`

## 4. Verification

- [x] 4.1 Re-run the Morph frontend unit tests
- [x] 4.2 Run the Morph frontend production build
