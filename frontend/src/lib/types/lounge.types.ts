export type LoungeStationType = "pc" | "cab" | "check_in";

export type LoungeStation = {
  id: string;
  label: string;
  station_type: LoungeStationType;
  closed: boolean;
  row: number;
  col: number;
  width: number;
  height: number;
  active_session_id: string | null;
  active_session_check_in_time: string | null;
  in_use: boolean;
};

export type LoungeStationStatus = "available" | "in_use" | "closed";

export type LoungeSessionDetails = {
  id: string;
  station_id: string;
  membership_id: string | null;
  membership_tier_title: string | null;
  check_in_time: string;
  check_out_time: string | null;
  game: string | null;
  user_id: string;
  user_full_name: string;
  user_email: string;
  user_student_id: string | null;
  user_avatar_url: string | null;
  checked_in_by: string;
  checked_in_by_name: string;
};

export type LoungeSession = {
  id: string;
  user_id: string;
  membership_id: string | null;
  station_id: string;
  check_in_time: string;
  check_out_time: string | null;
  checked_in_by: string;
  checked_out_by: string | null;
  game: string | null;
};
