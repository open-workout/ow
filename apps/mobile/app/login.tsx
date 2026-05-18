import { View, Text, TouchableOpacity, StatusBar, ActivityIndicator } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import { useAuth } from './context/auth';

export default function LoginScreen() {
  const router = useRouter();
  const { login } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleLogin = async () => {
    setLoading(true);
    setError('');
    try {
      await login();
      router.replace('/home');
    } catch (e: any) {
      if (e?.error !== 'a0.session.user_cancelled') {
        setError('Sign in failed. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: '#0a0a0a' }} edges={['top']}>
      <StatusBar barStyle="light-content" backgroundColor="#0a0a0a" />

      <View style={{ paddingHorizontal: 24, paddingTop: 8 }}>
        <TouchableOpacity
          onPress={() => router.back()}
          style={{ width: 40, height: 40, borderRadius: 20, backgroundColor: '#18181b', alignItems: 'center', justifyContent: 'center', borderWidth: 1, borderColor: '#27272a' }}
        >
          <Ionicons name="arrow-back" size={20} color="#a1a1aa" />
        </TouchableOpacity>
      </View>

      <View style={{ paddingHorizontal: 24, paddingTop: 40, flex: 1 }}>
        <Text style={{ color: '#fff', fontSize: 28, fontWeight: '800', letterSpacing: -0.5, marginBottom: 8 }}>
          Welcome back
        </Text>
        <Text style={{ color: '#71717a', fontSize: 16, marginBottom: 40, lineHeight: 24 }}>
          Sign in to continue tracking your progress.
        </Text>

        {error !== '' && (
          <View style={{ backgroundColor: 'rgba(239,68,68,0.07)', borderWidth: 1, borderColor: 'rgba(239,68,68,0.2)', borderRadius: 14, padding: 16, marginBottom: 24 }}>
            <Text style={{ color: '#ef4444', fontSize: 14 }}>{error}</Text>
          </View>
        )}

        <TouchableOpacity activeOpacity={0.85} onPress={handleLogin} disabled={loading}>
          <LinearGradient
            colors={['#f4f4f5', '#a1a1aa']}
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
            style={{ borderRadius: 16, paddingVertical: 17, alignItems: 'center' }}
          >
            {loading
              ? <ActivityIndicator color="#09090b" />
              : <Text style={{ color: '#09090b', fontWeight: '700', fontSize: 17 }}>Sign In</Text>
            }
          </LinearGradient>
        </TouchableOpacity>

        <View style={{ flexDirection: 'row', justifyContent: 'center', marginTop: 32 }}>
          <Text style={{ color: '#71717a', fontSize: 14 }}>Don't have an account? </Text>
          <TouchableOpacity onPress={() => router.replace('/signup')}>
            <Text style={{ color: '#f4f4f5', fontSize: 14, fontWeight: '600' }}>Create one</Text>
          </TouchableOpacity>
        </View>
      </View>
    </SafeAreaView>
  );
}
