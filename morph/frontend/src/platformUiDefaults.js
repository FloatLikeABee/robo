/** Keys and defaults must stay in sync with handlers/tran_platform_ui_config.go */
export const defaultPlatformUILabels = {
  product_name: 'MorphData',
  ai_assistant_name: 'Morph AI',
  nav_districts_facilities: 'Places',
  nav_people: 'People',
  nav_assets: 'Assets',
  nav_activities: 'Activities',
  nav_generic_data: 'Generic data',
  nav_big_notes: 'Big notes',
  nav_user_settings: 'User settings',
  nav_display_labels: 'Display names',
  term_facility: 'Place',
  term_facilities: 'Places',
  col_facility_code: 'Code',
  col_facility_name: 'Name',
  col_facility_type: 'Type',
  empty_districts_facilities: 'No places',
  col_activity_days: 'Activity days',
  col_linked_asset_id: 'Asset ID',
};

/** Keys per card on the Display names settings page (must cover every key in defaultPlatformUILabels). */
export const displayLabelSettingsGroups = {
  brandingAndNavigation: [
    'product_name',
    'ai_assistant_name',
    'nav_districts_facilities',
    'nav_people',
    'nav_assets',
    'nav_activities',
    'nav_generic_data',
    'nav_big_notes',
    'nav_user_settings',
    'nav_display_labels',
  ],
  termsAndTableColumns: [
    'term_facility',
    'term_facilities',
    'col_facility_code',
    'col_facility_name',
    'col_facility_type',
    'empty_districts_facilities',
    'col_activity_days',
    'col_linked_asset_id',
  ],
};

/** Short descriptions for the Display names settings form */
export const platformUILabelHints = {
  product_name: 'App name in the sidebar and footer',
  ai_assistant_name: 'Name shown on the AI chat panel',
  nav_districts_facilities: 'Work data → Places',
  nav_people: 'Work data → People (members + employees)',
  nav_assets: 'Assets (menu + grid title)',
  nav_activities: 'Activities (menu + grid title)',
  nav_generic_data: 'Generic data (Work data — CSV/JSON/PDF imports)',
  nav_big_notes: 'Big notes — AI HTML/MD notes with ComposerX publish',
  nav_user_settings: 'Configuration → users',
  nav_display_labels: 'Configuration → this screen',
  term_facility: 'Singular place label (columns, selectors)',
  term_facilities: 'Plural places label',
  col_facility_code: 'Places table — code column',
  col_facility_name: 'Places table — name column',
  col_facility_type: 'Places table — type column',
  empty_districts_facilities: 'Empty state on places table',
  col_activity_days: 'Activities grid — days column',
  col_linked_asset_id: 'Activities grid — linked asset',
};

/** Defaults match handlers/platform_config_v2.go defaultEntityDictionaries */
export const defaultEntityDictionaries = {
  employ_type: [
    { code: 'full_time', label: 'Full time' },
    { code: 'part_time', label: 'Part time' },
    { code: 'contractor', label: 'Contractor' },
  ],
  asset_type: [
    { code: 'bus', label: 'Bus' },
    { code: 'van', label: 'Van' },
    { code: 'car', label: 'Car' },
    { code: 'other', label: 'Other' },
  ],
  activity_type: [
    { code: 'standard', label: 'Standard route' },
    { code: 'field_trip', label: 'Field trip' },
    { code: 'sports', label: 'Sports / events' },
  ],
  facility_type: [
    { code: 'school', label: 'School' },
    { code: 'depot', label: 'Depot / yard' },
    { code: 'office', label: 'Office' },
  ],
  participant_type: [
    { code: 'student', label: 'Student' },
    { code: 'passenger', label: 'Passenger' },
    { code: 'chaperone', label: 'Chaperone' },
  ],
};
