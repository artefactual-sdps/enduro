# Plan-First Development Workflow

You MUST follow this workflow for any feature development or significant changes:

## Step 1: Analysis & Planning
- Analyze the current state and requirements
- Research the codebase to understand existing patterns
- Design the solution architecture
- Create a detailed implementation plan

## Step 2: Save Plan
- **First**, check existing plans in `.claude/plans/` to determine the next sequential number
- Save the plan as a markdown file in `.claude/plans/[NNNN_feature_name].md`
- Use enumerated filenames with underscore separators (e.g., `0001_user_authentication.md`)
- Use descriptive names that match the feature being implemented
- **IMPORTANT**: Include the original user input at the top of the plan file in a "User Request" section
- Add a "Related Files" section referencing the future implementation file: `.claude/implms/[NNNN_feature_name].md`
- Include all technical details, file changes, and implementation steps

## Step 3: Present Plan
- Use the `exit_plan_mode` tool to present the plan to the user
- Wait for user approval before proceeding with implementation
- Do NOT start coding until the plan is approved

## Step 4: Implementation
- Follow the approved plan step by step
- Use the TodoWrite tool to track progress throughout implementation
- Make changes according to the plan

## Step 5: Documentation
- Save implementation details in `.claude/implms/[NNNN_feature_name].md`
- Use the same enumerated filename as the plan with underscore separators
- Add a "Related Files" section referencing the original plan file: `.claude/plans/[NNNN_feature_name].md`
- Document what was actually implemented vs. planned
- Include validation results and any deviations from the plan

## Important Reminders
- NEVER skip the planning phase
- ALWAYS get user approval before implementation
- ALWAYS document both plans and results
- Use enumerated filenames with underscores (not hyphens) for both plan and implementation docs
- ALWAYS check existing plan numbers to use the next sequential number
- ALWAYS include cross-references between plans and implementations