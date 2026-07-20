import { z } from 'zod';

const restriction = z.object({
	day: z.number().int().min(0).max(6),
	start: z.number().min(0).max(24),
	end: z.number().min(0).max(24)
});

export const layerSchema = z.object({
	id: z.string().min(1),
	participants: z.array(z.string()).min(1, 'A layer needs at least one participant.'),
	rotation: z.enum(['daily', 'weekly', 'custom']),
	intervalDays: z.number().int().min(1, 'At least one day.').max(30, 'At most 30 days.'),
	handoverHour: z.number().int().min(0).max(23),
	startsOn: z.iso.date('Pick a start date.'),
	restrictions: z.array(restriction)
});

export const scheduleSchema = z.object({
	name: z
		.string()
		.min(1, 'Give the schedule a name.')
		.regex(
			/^[a-z][a-z0-9-]{0,39}$/,
			'Start with a letter; lower case letters, numbers and dashes — it is used in the feed URL.'
		)
		// these collide with route segments and the preview endpoint
		.refine((name) => !['new', 'mine', 'preview'].includes(name), 'That name is reserved. Pick another.'),
	team: z.string().min(1, 'Pick a team.'),
	timezone: z.string().min(1, 'Pick a timezone.'),
	layers: z.array(layerSchema).min(1, 'A schedule needs at least one layer.')
});

export type ScheduleSchema = typeof scheduleSchema;
