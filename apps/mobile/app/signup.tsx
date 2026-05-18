import {
  View, Text, TextInput, TouchableOpacity, ScrollView,
  StatusBar, KeyboardAvoidingView, Platform, ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useState, useMemo } from 'react';
import FormField from './components/FormField';
import { useAuth } from './context/auth';

// ─── Sports data ────────────────────────────────────────────────────────────

const SPORTS_CATEGORIES = [
  {
    title: 'Strength & Power',
    sports: ['Bodybuilding', 'Weightlifting', 'Powerlifting', 'CrossFit', 'Strongman', 'Calisthenics', 'Olympic Lifting', 'Kettlebell Sport'],
  },
  {
    title: 'Combat Sports',
    sports: ['Boxing', 'MMA', 'Wrestling', 'Judo', 'Brazilian Jiu-Jitsu', 'Karate', 'Muay Thai', 'Taekwondo', 'Kickboxing', 'Fencing', 'Sambo', 'Kung Fu', 'Capoeira', 'Savate'],
  },
  {
    title: 'Team Sports',
    sports: ['Soccer', 'Basketball', 'American Football', 'Baseball', 'Rugby', 'Rugby League', 'Volleyball', 'Ice Hockey', 'Cricket', 'Handball', 'Lacrosse', 'Water Polo', 'Field Hockey', 'Australian Football', 'Futsal', 'Beach Volleyball', 'Netball', 'Softball', 'Floorball', 'Kabaddi'],
  },
  {
    title: 'Racket Sports',
    sports: ['Tennis', 'Badminton', 'Squash', 'Table Tennis', 'Padel', 'Pickleball', 'Racquetball', 'Racketball'],
  },
  {
    title: 'Athletics & Running',
    sports: ['Sprinting', 'Long Distance Running', 'Marathon', 'Trail Running', 'Ultra Running', 'Hurdles', 'High Jump', 'Long Jump', 'Triple Jump', 'Shot Put', 'Javelin', 'Discus', 'Hammer Throw', 'Pole Vault', 'Decathlon', 'Heptathlon', 'Triathlon', 'Duathlon', 'Racewalking', 'Obstacle Racing'],
  },
  {
    title: 'Water Sports',
    sports: ['Swimming', 'Open Water Swimming', 'Diving', 'Surfing', 'Rowing', 'Kayaking', 'Canoeing', 'Sailing', 'Water Skiing', 'Windsurfing', 'Kitesurfing', 'Stand-Up Paddleboarding', 'Synchronized Swimming', 'Wakeboarding', 'Freediving'],
  },
  {
    title: 'Cycling',
    sports: ['Road Cycling', 'Mountain Biking', 'BMX', 'Track Cycling', 'Gravel Cycling', 'Cyclocross', 'Downhill Cycling', 'Bike Trials'],
  },
  {
    title: 'Winter Sports',
    sports: ['Alpine Skiing', 'Snowboarding', 'Ice Skating', 'Speed Skating', 'Figure Skating', 'Biathlon', 'Cross-Country Skiing', 'Ski Jumping', 'Bobsled', 'Luge', 'Skeleton', 'Curling', 'Freestyle Skiing', 'Ice Climbing'],
  },
  {
    title: 'Individual & Other',
    sports: ['Golf', 'Gymnastics', 'Artistic Gymnastics', 'Rhythmic Gymnastics', 'Trampoline', 'Rock Climbing', 'Bouldering', 'Archery', 'Shooting', 'Equestrian', 'Skateboarding', 'Yoga', 'Pilates', 'Dance', 'Cheerleading', 'Parkour', 'Motocross', 'Formula Racing', 'Roller Derby', 'Esports', 'Weightlifting'],
  },
];

const GENDERS = ['Male', 'Female', 'Non-binary', 'Prefer not to say'];
const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080';

// ─── Main component ─────────────────────────────────────────────────────────

export default function SignupScreen() {
  const router = useRouter();
  const { login, accessToken, userID } = useAuth();
  // step 1 = Auth0 sign-up, step 2 = profile, step 3 = sports
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [authLoading, setAuthLoading] = useState(false);
  const [submitLoading, setSubmitLoading] = useState(false);
  const [authError, setAuthError] = useState('');

  // Step 2 — profile
  const [username, setUsername] = useState('');
  const [age, setAge] = useState(25);
  const [gender, setGender] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Step 3 — sports
  const [selectedSports, setSelectedSports] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState('');

  const clearError = (key: string) => setErrors((prev) => ({ ...prev, [key]: '' }));

  const handleBack = () => {
    if (step === 1) router.back();
    else setStep((prev) => (prev - 1) as 1 | 2);
  };

  const handleAuth0Signup = async () => {
    setAuthLoading(true);
    setAuthError('');
    try {
      await login();
      setStep(2);
    } catch (e: any) {
      if (e?.error !== 'a0.session.user_cancelled') {
        setAuthError('Sign up failed. Please try again.');
      }
    } finally {
      setAuthLoading(false);
    }
  };

  const continueFromStep2 = () => {
    const errs: Record<string, string> = {};
    if (!username.trim() || username.length < 3) errs.username = 'Must be at least 3 characters';
    if (/\s/.test(username)) errs.username = 'No spaces allowed';
    if (age < 13 || age > 99) errs.age = 'Please enter a valid age (13–99)';
    if (!gender) errs.gender = 'Please select your gender';

    if (Object.keys(errs).length > 0) { setErrors(errs); return; }
    setErrors({});
    setStep(3);
  };

  const handleGetStarted = async () => {
    if (!accessToken || !userID) return;
    setSubmitLoading(true);
    try {
      const birthYear = new Date().getFullYear() - age;
      await fetch(`${API_URL}/users/`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          username,
          gender,
          birthdate: `${birthYear}-01-01`,
          sport_goals: selectedSports,
        }),
      });
      router.replace('/home');
    } catch {
      router.replace('/home');
    } finally {
      setSubmitLoading(false);
    }
  };

  const toggleSport = (sport: string) => {
    setSelectedSports((prev) =>
      prev.includes(sport) ? prev.filter((s) => s !== sport) : prev.length < 3 ? [...prev, sport] : prev
    );
  };

  const filteredCategories = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return SPORTS_CATEGORIES;
    return SPORTS_CATEGORIES
      .map((cat) => ({ ...cat, sports: cat.sports.filter((s) => s.toLowerCase().includes(q)) }))
      .filter((cat) => cat.sports.length > 0);
  }, [searchQuery]);

  const stepTitles = ['Create Account', 'About You', 'Your Sports'];
  const stepSubtitles = [
    'Sign up securely to get started.',
    'Help us personalize your experience.',
    'Select up to 3 sports you train for.',
  ];

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: '#0a0a0a' }} edges={['top']}>
      <StatusBar barStyle="light-content" backgroundColor="#0a0a0a" />

      {/* ── Fixed header ── */}
      <View style={{ paddingHorizontal: 24, paddingTop: 8 }}>
        <TouchableOpacity
          onPress={handleBack}
          style={{ width: 40, height: 40, borderRadius: 20, backgroundColor: '#18181b', alignItems: 'center', justifyContent: 'center', borderWidth: 1, borderColor: '#27272a', marginBottom: 24 }}
        >
          <Ionicons name="arrow-back" size={20} color="#a1a1aa" />
        </TouchableOpacity>

        <View style={{ flexDirection: 'row', gap: 8, marginBottom: 28 }}>
          {[1, 2, 3].map((s) => (
            <View key={s} style={{ flex: 1, height: 3, borderRadius: 2, backgroundColor: s <= step ? '#f4f4f5' : '#27272a' }} />
          ))}
        </View>

        <Text style={{ color: '#fff', fontSize: 26, fontWeight: '800', letterSpacing: -0.5, marginBottom: 6 }}>
          {stepTitles[step - 1]}
        </Text>
        <Text style={{ color: '#71717a', fontSize: 15, marginBottom: 24 }}>
          {stepSubtitles[step - 1]}
        </Text>
      </View>

      {/* ── Step 1: Auth0 sign-up ── */}
      {step === 1 && (
        <View style={{ paddingHorizontal: 24, flex: 1 }}>
          {authError !== '' && (
            <View style={{ backgroundColor: 'rgba(239,68,68,0.07)', borderWidth: 1, borderColor: 'rgba(239,68,68,0.2)', borderRadius: 14, padding: 16, marginBottom: 24 }}>
              <Text style={{ color: '#ef4444', fontSize: 14 }}>{authError}</Text>
            </View>
          )}

          <TouchableOpacity activeOpacity={0.85} onPress={handleAuth0Signup} disabled={authLoading} style={{ marginTop: 8 }}>
            <LinearGradient
              colors={['#f4f4f5', '#a1a1aa']}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
              style={{ borderRadius: 16, paddingVertical: 17, alignItems: 'center' }}
            >
              {authLoading
                ? <ActivityIndicator color="#09090b" />
                : <Text style={{ color: '#09090b', fontWeight: '700', fontSize: 17 }}>Continue</Text>
              }
            </LinearGradient>
          </TouchableOpacity>

          <View style={{ flexDirection: 'row', justifyContent: 'center', marginTop: 24 }}>
            <Text style={{ color: '#71717a', fontSize: 14 }}>Already have an account? </Text>
            <TouchableOpacity onPress={() => router.replace('/login')}>
              <Text style={{ color: '#f4f4f5', fontSize: 14, fontWeight: '600' }}>Sign in</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}

      {/* ── Step 2: Profile ── */}
      {step === 2 && (
        <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : 'height'} style={{ flex: 1 }}>
          <ScrollView
            style={{ flex: 1 }}
            contentContainerStyle={{ paddingHorizontal: 24, paddingBottom: 48 }}
            keyboardShouldPersistTaps="handled"
            showsVerticalScrollIndicator={false}
          >
            <FormField
              label="Username *"
              value={username}
              onChangeText={(t) => { setUsername(t); clearError('username'); }}
              placeholder="Choose a username"
              autoCapitalize="none"
              autoCorrect={false}
              error={errors.username}
            />

            {/* Age stepper */}
            <Text style={{ color: '#a1a1aa', fontSize: 12, fontWeight: '600', marginBottom: 16, textTransform: 'uppercase', letterSpacing: 0.8 }}>
              Age *
            </Text>
            <View style={{ flexDirection: 'row', alignItems: 'center', gap: 24, marginBottom: 6 }}>
              <TouchableOpacity
                onPress={() => setAge((a) => Math.max(13, a - 1))}
                style={{ width: 48, height: 48, borderRadius: 12, backgroundColor: '#18181b', borderWidth: 1, borderColor: '#27272a', alignItems: 'center', justifyContent: 'center' }}
              >
                <Ionicons name="remove" size={22} color="#a1a1aa" />
              </TouchableOpacity>
              <Text style={{ color: '#fff', fontSize: 36, fontWeight: '700', minWidth: 64, textAlign: 'center' }}>
                {age}
              </Text>
              <TouchableOpacity
                onPress={() => setAge((a) => Math.min(99, a + 1))}
                style={{ width: 48, height: 48, borderRadius: 12, backgroundColor: '#18181b', borderWidth: 1, borderColor: '#27272a', alignItems: 'center', justifyContent: 'center' }}
              >
                <Ionicons name="add" size={22} color="#a1a1aa" />
              </TouchableOpacity>
            </View>
            {errors.age && <Text style={{ color: '#ef4444', fontSize: 12, marginBottom: 8 }}>{errors.age}</Text>}

            <View style={{ height: 1, backgroundColor: '#18181b', marginVertical: 32 }} />

            {/* Gender selector */}
            <Text style={{ color: '#a1a1aa', fontSize: 12, fontWeight: '600', marginBottom: 16, textTransform: 'uppercase', letterSpacing: 0.8 }}>
              Gender *
            </Text>
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 10 }}>
              {GENDERS.map((g) => {
                const active = gender === g;
                return (
                  <TouchableOpacity
                    key={g}
                    onPress={() => { setGender(g); clearError('gender'); }}
                    style={{
                      paddingHorizontal: 22, paddingVertical: 13, borderRadius: 12,
                      backgroundColor: active ? '#f4f4f5' : '#18181b',
                      borderWidth: 1, borderColor: active ? '#f4f4f5' : '#27272a',
                    }}
                  >
                    <Text style={{ color: active ? '#09090b' : '#d4d4d8', fontWeight: active ? '700' : '400', fontSize: 15 }}>
                      {g}
                    </Text>
                  </TouchableOpacity>
                );
              })}
            </View>
            {errors.gender && <Text style={{ color: '#ef4444', fontSize: 12, marginTop: 8 }}>{errors.gender}</Text>}

            <TouchableOpacity activeOpacity={0.85} onPress={continueFromStep2} style={{ marginTop: 40 }}>
              <LinearGradient
                colors={['#f4f4f5', '#a1a1aa']}
                start={{ x: 0, y: 0 }}
                end={{ x: 1, y: 1 }}
                style={{ borderRadius: 16, paddingVertical: 17, alignItems: 'center' }}
              >
                <Text style={{ color: '#09090b', fontWeight: '700', fontSize: 17 }}>Continue</Text>
              </LinearGradient>
            </TouchableOpacity>
          </ScrollView>
        </KeyboardAvoidingView>
      )}

      {/* ── Step 3: Sports ── */}
      {step === 3 && (
        <View style={{ flex: 1 }}>
          <View style={{ paddingHorizontal: 24, marginBottom: 4 }}>
            {selectedSports.length > 0 && (
              <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginBottom: 14 }}>
                {selectedSports.map((s) => (
                  <TouchableOpacity
                    key={s}
                    onPress={() => toggleSport(s)}
                    style={{ flexDirection: 'row', alignItems: 'center', gap: 6, backgroundColor: '#27272a', paddingHorizontal: 14, paddingVertical: 8, borderRadius: 999, borderWidth: 1, borderColor: '#3f3f46' }}
                  >
                    <Text style={{ color: '#f4f4f5', fontSize: 13, fontWeight: '600' }}>{s}</Text>
                    <Ionicons name="close" size={13} color="#71717a" />
                  </TouchableOpacity>
                ))}
              </View>
            )}

            <View style={{ flexDirection: 'row', alignItems: 'center', backgroundColor: '#18181b', borderWidth: 1, borderColor: '#27272a', borderRadius: 12, paddingHorizontal: 14, paddingVertical: 12, marginBottom: 10 }}>
              <Ionicons name="search-outline" size={18} color="#52525b" style={{ marginRight: 10 }} />
              <TextInput
                placeholder="Search sports..."
                placeholderTextColor="#52525b"
                value={searchQuery}
                onChangeText={setSearchQuery}
                style={{ flex: 1, color: '#fff', fontSize: 15 }}
              />
              {searchQuery.length > 0 && (
                <TouchableOpacity onPress={() => setSearchQuery('')}>
                  <Ionicons name="close-circle" size={18} color="#52525b" />
                </TouchableOpacity>
              )}
            </View>

            <Text style={{ color: '#52525b', fontSize: 13 }}>
              {selectedSports.length} / 3 selected
            </Text>
          </View>

          <ScrollView
            style={{ flex: 1 }}
            contentContainerStyle={{ paddingHorizontal: 24, paddingBottom: 120 }}
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
          >
            {filteredCategories.map((category) => (
              <View key={category.title} style={{ marginTop: 24 }}>
                <Text style={{ color: '#52525b', fontSize: 11, fontWeight: '700', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 12 }}>
                  {category.title}
                </Text>
                <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
                  {category.sports.map((sport) => {
                    const selected = selectedSports.includes(sport);
                    const disabled = selectedSports.length >= 3 && !selected;
                    return (
                      <TouchableOpacity
                        key={sport}
                        onPress={() => toggleSport(sport)}
                        disabled={disabled}
                        style={{
                          paddingHorizontal: 16, paddingVertical: 10, borderRadius: 999,
                          backgroundColor: selected ? '#f4f4f5' : '#18181b',
                          borderWidth: 1,
                          borderColor: selected ? '#f4f4f5' : disabled ? '#1c1c1e' : '#27272a',
                          opacity: disabled ? 0.35 : 1,
                        }}
                      >
                        <Text style={{ color: selected ? '#09090b' : '#d4d4d8', fontSize: 14, fontWeight: selected ? '600' : '400' }}>
                          {sport}
                        </Text>
                      </TouchableOpacity>
                    );
                  })}
                </View>
              </View>
            ))}
          </ScrollView>

          <View style={{ position: 'absolute', bottom: 0, left: 0, right: 0, paddingHorizontal: 24, paddingBottom: 36, paddingTop: 16, backgroundColor: '#0a0a0a', borderTopWidth: 1, borderTopColor: '#18181b' }}>
            <TouchableOpacity activeOpacity={0.85} onPress={handleGetStarted} disabled={submitLoading}>
              <LinearGradient
                colors={['#f4f4f5', '#a1a1aa']}
                start={{ x: 0, y: 0 }}
                end={{ x: 1, y: 1 }}
                style={{ borderRadius: 16, paddingVertical: 17, alignItems: 'center' }}
              >
                {submitLoading ? <ActivityIndicator color="#09090b" /> : (
                  <Text style={{ color: '#09090b', fontWeight: '700', fontSize: 17 }}>
                    {selectedSports.length > 0
                      ? `Get Started · ${selectedSports.length} sport${selectedSports.length > 1 ? 's' : ''} selected`
                      : 'Get Started'}
                  </Text>
                )}
              </LinearGradient>
            </TouchableOpacity>
          </View>
        </View>
      )}
    </SafeAreaView>
  );
}
